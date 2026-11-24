package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal"
	"github.com/soerenschneider/dyndns/v2/internal/client"
	"github.com/soerenschneider/dyndns/v2/internal/conf"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
	"github.com/soerenschneider/dyndns/v2/internal/server"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
	"golang.org/x/term"
)

const defaultConfigPath = "/etc/dyndns/config.json"
const notificationTopic = "dyndns/+"

type dyndnsApp interface {
	StartApp(ctx context.Context) error
}

type appServerMode struct {
	requests  chan update.UpdateRecordRequest
	server    *server.DyndnsServer
	listeners []listener
}

type appEdgeMode struct {
	serverApp dyndnsApp
	clientApp dyndnsApp
}

func (app appEdgeMode) StartApp(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		if err := app.serverApp.StartApp(ctx); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := app.clientApp.StartApp(ctx); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (app appServerMode) StartApp(ctx context.Context) error {
	log.Info().Str("component", "server").Msg("Ready, listening for incoming requests")

	wg := &sync.WaitGroup{}
	for _, listener := range app.listeners {
		go func() {
			log.Info().Str("component", "server").Msg("starting listener")
			err := listener.Listen(ctx, wg)
			if err != nil {
				log.Fatal().Err(err).Str("component", "server").Msg("could not start listener")
			}
		}()
	}

	app.server.Listen(ctx)
	close(app.requests)

	return nil
}

type appClientMode struct {
	client     *client.Client
	reconciler *client.Reconciler
	once       bool
}

func (app appClientMode) StartApp(ctx context.Context) error {
	// The update is dispatched synchronously, so there is nothing left to
	// reconcile in the background.
	if app.once {
		_, err := app.client.Resolve(ctx, nil)
		return err
	}

	go app.reconciler.Run(ctx)
	app.client.Run(ctx)

	return nil
}

func main() {
	metrics.Version.WithLabelValues(internal.BuildVersion, internal.CommitHash, internal.GoVersion).Set(1)
	metrics.ProcessStartTime.SetToCurrentTime()

	configPath := flag.String("config", defaultConfigPath, "Path to the config file")
	version := flag.Bool("version", false, "Print version and exit")
	debug := flag.Bool("debug", false, "Print debug logs")
	once := flag.Bool("once", false, "Resolve and update the record once, then exit (client mode only)")
	forceSendUpdate := flag.Bool("force", false, "Force sending an update request at start")
	genKeypair := flag.Bool("gen-keypair", false, "Generate a keypair, print it and exit")
	flag.Parse()

	if *version {
		fmt.Printf("%s (commit: %s) go%s\n", internal.BuildVersion, internal.CommitHash, internal.GoVersion)
		os.Exit(0)
	}

	if *genKeypair {
		generateKeypair()
	}

	initLogging(*debug)

	metrics.Version.WithLabelValues(internal.BuildVersion, internal.CommitHash, internal.GoVersion).Set(1)
	metrics.ProcessStartTime.SetToCurrentTime()

	config, err := conf.ReadConfig(*configPath)
	if err != nil {
		if *configPath != defaultConfigPath {
			dieOnError(err, "couldn't read config file")
		}
		config = conf.GetDefaultConfig()
	}

	err = conf.ParseEnvVariables(config)
	dieOnError(err, "could not parse env variables")

	err = conf.ValidateConfig(config)
	dieOnError(err, "Config validation failed")

	err = applyCliFlags(config, *once, *forceSendUpdate)
	dieOnError(err, "invalid flags")

	ctx := context.Background()

	if !*once {
		go func() {
			err := metrics.StartMetricsServer(ctx, config.MetricsListener)
			dieOnError(err, "could not start metrics server")
		}()
		go metrics.StartHeartbeat(ctx)
	}

	app, err := buildApp(config)
	dieOnError(err, "could not build dependencies")

	run(ctx, app)
}

// applyCliFlags checks the cli flags against the configured mode and stores
// them in the config.
func applyCliFlags(config *conf.Conf, once, forceSendUpdate bool) error {
	if once && config.Mode != internal.ClientMode {
		return errors.New("-once is only supported in client mode")
	}

	if forceSendUpdate && config.Mode == internal.ServerMode {
		return errors.New("-force is not supported in server mode")
	}

	config.Client.Once = once
	config.Client.ForceSendUpdate = forceSendUpdate
	return nil
}

func buildApp(config *conf.Conf) (dyndnsApp, error) {
	switch config.Mode {
	case internal.EdgeMode:
		return newEdge(config)
	case internal.ClientMode:
		dispatchers, err := buildUpdateDispatchers(config)
		if err != nil {
			return nil, err
		}
		if len(dispatchers) == 0 {
			return nil, errors.New("mode is client and no dispatchers were built")
		}
		return newClient(config, dispatchers)
	case internal.ServerMode:
		return newServer(config)
	}

	return nil, fmt.Errorf("unknown mode %q", config.Mode)
}

// run starts app and blocks until it returns, e.g. after a single run with
// -once, or until a termination signal is caught.
func run(ctx context.Context, app dyndnsApp) {
	done := make(chan error, 1)
	go func() {
		done <- app.StartApp(ctx)
	}()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-done:
		dieOnError(err, "error running dyndns")
	case <-terminate:
		log.Info().Str("component", "server").Msg("Caught signal, cancelling context")
	}
}

func dieOnError(err error, msg string) {
	if err != nil {
		log.Fatal().Err(err).Msg(msg)
	}
}

type listener interface {
	Listen(ctx context.Context, wg *sync.WaitGroup) error
}

func initLogging(debug bool) {
	//#nosec:G115
	if term.IsTerminal(int(os.Stdout.Fd())) {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		})
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Msgf("Started dyndns version %s, commit %s", internal.BuildVersion, internal.CommitHash)
}
