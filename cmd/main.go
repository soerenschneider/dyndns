package main

import (
	"context"
	"errors"
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
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func buildApp(ctx context.Context, config *conf.Conf) (dyndnsApp, error) {
	switch config.Mode {
	case internal.EdgeMode:
		return newEdge(ctx, config)
	case internal.ClientMode:
		dispatchers, err := buildUpdateDispatchers(ctx, config)
		if err != nil {
			return nil, err
		}
		if len(dispatchers) == 0 {
			return nil, errors.New("mode is client and no dispatchers were built")
		}
		return newClient(ctx, config, dispatchers)
	case internal.ServerMode:
		return newServer(ctx, config)
	}

	return nil, fmt.Errorf("unknown mode %q", config.Mode)
}

// run starts app and blocks until it returns, e.g. after a single run with
// --once, or until a termination signal is caught.
func run(ctx context.Context, app dyndnsApp) error {
	done := make(chan error, 1)
	go func() {
		done <- app.StartApp(ctx)
	}()

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("error running dyndns: %w", err)
		}
	case <-terminate:
		log.Info().Str("component", "server").Msg("Caught signal, cancelling context")
	}

	return nil
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
