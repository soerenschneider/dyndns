package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal"
	"github.com/soerenschneider/dyndns/v2/internal/client"
	"github.com/soerenschneider/dyndns/v2/internal/client/dispatchers"
	"github.com/soerenschneider/dyndns/v2/internal/client/resolvers"
	"github.com/soerenschneider/dyndns/v2/internal/conf"
	"github.com/soerenschneider/dyndns/v2/internal/events/mqtt"
	sink "github.com/soerenschneider/dyndns/v2/internal/events/nats"
	"github.com/soerenschneider/dyndns/v2/internal/notification"
	"github.com/soerenschneider/dyndns/v2/internal/util"
	"github.com/soerenschneider/dyndns/v2/internal/verification"
	"github.com/soerenschneider/dyndns/v2/internal/verification/key_provider"
	"go.uber.org/multierr"
)

func newClient(config *conf.Conf, updaters map[string]client.RecordUpdater) (*appClientMode, error) {
	var keypair verification.SignatureKeypair
	switch config.Mode {
	case internal.EdgeMode:
		keypair = &verification.Edge{}
	case internal.ClientMode:
		provider, err := buildKeyProvider(config.Client)
		if err != nil {
			return nil, err
		}

		keypair, err = getKeypair(provider)
		if err != nil {
			return nil, err
		}
	}

	notificationImpl, err := buildNotificationImpl(config)
	if err != nil {
		return nil, err
	}

	resolver, err := buildResolver(config.Client)
	if err != nil {
		return nil, err
	}

	reconciler, err := client.NewReconciler(updaters, true)
	if err != nil {
		return nil, err
	}

	opts := []client.Opts{
		client.WithInterval(15 * time.Second),
	}

	client, err := client.NewClient(resolver, keypair, reconciler, notificationImpl, opts...)
	if err != nil {
		return nil, err
	}

	return &appClientMode{
		client:     client,
		reconciler: reconciler,
	}, nil
}

// nolint cyclop
func buildUpdateDispatchers(config *conf.Conf) (map[string]client.RecordUpdater, error) {
	dispatchImplementations := map[string]client.RecordUpdater{}

	var errs error
	if len(config.Client.Brokers) > 0 {
		log.Info().Str("component", "client").Msg("Building MQTT notifier(s)")
		for _, broker := range config.Client.Brokers {
			dispatcher, err := mqtt.NewMqttClient(broker, config.Client.ClientId, fmt.Sprintf("dyndns/%s", config.Client.Host), config.Client.TlsConfig())
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				dispatchImplementations[broker] = dispatcher
			}
		}
	}

	if config.Client.IsConfiguredForUpdates() {
		log.Info().Msg("Building NATS notifier")
		js, err := sink.Connect(config.Client.NatsConfig)
		if err != nil {
			errs = multierr.Append(errs, err)
		} else {
			dispatcher, err := sink.NewNatsDyndnsClient(&config.Client.NatsConfig, js)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				dispatchImplementations[config.Client.Url] = dispatcher
			}
		}
	}

	if len(config.Client.HttpDispatcherConf) > 0 {
		log.Info().Str("component", "client").Msg("Building HTTP notifier")
		for _, dispatcher := range config.Client.HttpDispatcherConf {
			httpDispatcher, err := dispatchers.NewHttpDispatcher(dispatcher.Url)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				dispatchImplementations[dispatcher.Url] = httpDispatcher
			}
		}
	}

	if len(config.Client.SqsQueue) > 0 {
		log.Info().Str("component", "client").Msg("Building AWS SQS notifier")
		sqs, err := dispatchers.NewSqsDispatcher(config.Client.SqsConfig, nil)
		if err != nil {
			errs = multierr.Append(errs, err)
		} else {
			dispatchImplementations["sqs"] = sqs
		}
	}

	return dispatchImplementations, errs
}

func buildResolver(conf *conf.ClientConf) (resolvers.IpResolver, error) {
	if len(conf.NetworkInterface) > 0 {
		log.Info().Str("component", "client").Msgf("Building new resolver for interface %s", conf.NetworkInterface)
		return resolvers.NewInterfaceResolver(conf.NetworkInterface, conf.Host)
	}

	log.Info().Str("component", "client").Msg("Building HTTP resolver")
	return resolvers.NewHttpResolver(conf.Host, conf.PreferredUrls, conf.FallbackUrls, conf.AddrFamilies)
}

func buildNotificationImpl(config *conf.Conf) (notification.Notification, error) {
	if config.BuildEmailNotification() {
		// TODO: fail fast
		return util.NewEmailNotification(&config.EmailConfig)
	}

	//if config.SupportsCloudeventsDispatch() {
	//	jetstream, err := sink.Connect(config.NatsConfig)
	//	dieOnError(err, "could not build nats jetstream")
	//	return sink.NewNatsCloudevents(&config.NatsConfig, jetstream)
	//}

	return &notification.DummyNotification{}, nil
}

func buildKeyProvider(config *conf.ClientConf) (key_provider.KeyProvider, error) {
	if len(config.KeyPair) > 0 {
		return key_provider.NewEnvProvider(config.KeyPair)
	}

	return key_provider.NewFileProvider(config.KeyPairPath)
}

func getKeypair(provider key_provider.KeyProvider) (verification.SignatureKeypair, error) {
	log.Info().Str("component", "client").Msg("Trying to read keypair")
	reader, err := provider.Reader()
	if err != nil {
		return nil, fmt.Errorf("could not acquire reader to read keypair: %w", err)
	}

	keypair, err := verification.FromReader(reader)
	if err == nil {
		return keypair, nil
	}

	if !provider.CanWrite() {
		return nil, fmt.Errorf("writer does not support creating a new keypair: %w", err)
	}

	log.Info().Err(err).Str("component", "client").Msg("Creating new keypair, existing keypair could not be read")
	keypair, err = verification.NewKeyPair()
	if err != nil {
		return nil, err
	}
	log.Info().Str("component", "client").Str("public_key", base64.StdEncoding.EncodeToString(keypair.PubKey)).Msg("Created new keypair")

	jsonData, err := keypair.AsJson()
	if err != nil {
		return nil, err
	}

	if err = provider.Write(jsonData); err != nil {
		return nil, fmt.Errorf("could not save keypair: %w", err)
	}

	return keypair, nil
}

func generateKeypair() {
	keypair, err := verification.NewKeyPair()
	if err != nil {
		log.Fatal().Str("component", "client").Err(err).Msg("Can not create keypair")
	}

	jsonEncoded, err := keypair.AsJson()
	if err != nil {
		log.Fatal().Str("component", "client").Err(err).Msg("could not marshall keypair")
	}
	fmt.Printf("%s\n", jsonEncoded)
	os.Exit(0)
}
