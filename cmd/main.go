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
}

func (app appClientMode) StartApp(ctx context.Context) error {
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
	flag.Parse()

	if *version {
		fmt.Printf("%s (commit: %s) go%s\n", internal.BuildVersion, internal.CommitHash, internal.GoVersion)
		os.Exit(0)
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

	ctx := context.Background()

	go func() {
		err := metrics.StartMetricsServer(ctx, config.MetricsListener)
		dieOnError(err, "could not start metrics server")
	}()
	go metrics.StartHeartbeat(ctx)

	var app dyndnsApp
	switch config.Mode {
	case internal.EdgeMode:
		app, err = newEdge(config)
	case internal.ClientMode:
		var dispatchers map[string]client.RecordUpdater
		dispatchers, err = buildUpdateDispatchers(config)
		if err != nil {
			break
		}
		if len(dispatchers) == 0 {
			err = errors.New("mode is client and no dispatchers were built")
			break
		}
		app, err = newClient(config, dispatchers)
	case internal.ServerMode:
		app, err = newServer(config)
	}

	dieOnError(err, "could not build dependencies")

	go func() {
		if err := app.StartApp(ctx); err != nil {
			log.Fatal().Err(err).Msg("error running dyndns")
		}
	}()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	<-terminate
	log.Info().Str("component", "server").Msg("Caught signal, cancelling context")
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
