package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal"
	"github.com/soerenschneider/dyndns/v2/internal/conf"
	"github.com/soerenschneider/dyndns/v2/internal/events/http"
	"github.com/soerenschneider/dyndns/v2/internal/events/mqtt"
	sink "github.com/soerenschneider/dyndns/v2/internal/events/nats"
	client "github.com/soerenschneider/dyndns/v2/internal/events/sqs"
	"github.com/soerenschneider/dyndns/v2/internal/server"
	"github.com/soerenschneider/dyndns/v2/internal/server/dns"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
	"go.uber.org/multierr"
)

func newServer(config *conf.Conf) (*appServerMode, error) {
	notificationImpl, err := buildNotificationImpl(config)
	if err != nil {
		return nil, err
	}

	var requestsChannel = make(chan update.UpdateRecordRequest)
	provider, err := buildAwsCredentialsProvider(config.Server)
	if err != nil {
		return nil, err
	}

	listeners, err := buildListeners(*config.Server, requestsChannel, provider)
	if err != nil {
		log.Error().Err(err).Msg("could not build all listeners")
	}
	if len(listeners) == 0 && config.Mode == internal.ServerMode {
		log.Fatal().Err(err).Msg("no usable listener has been built")
	}

	propagator, err := dns.NewRoute53Propagator(config.Server.HostedZoneId, provider)
	if err != nil {
		return nil, err
	}

	dyndnsServer, err := server.NewServer(*config, propagator, requestsChannel, notificationImpl)
	if err != nil {
		return nil, err
	}

	return &appServerMode{
		server:    dyndnsServer,
		listeners: listeners,
		requests:  requestsChannel,
	}, nil
}

func buildSqs(config conf.ServerConf, requests chan update.UpdateRecordRequest, credProvider aws.CredentialsProvider) (*client.SqsListener, error) {
	return client.NewSqsConsumer(config.SqsConfig, credProvider, requests)
}

func buildMqtt(config conf.ServerConf, requests chan update.UpdateRecordRequest) ([]*mqtt.MqttBus, error) {
	var servers []*mqtt.MqttBus
	for _, broker := range config.Brokers {
		mqttServer, err := mqtt.NewMqttServer(broker, config.ClientId, notificationTopic, config.TlsConfig(), requests)
		if err != nil {
			log.Error().Err(err).Str("component", "server").Msg("could not connect to mqtt")
		} else {
			servers = append(servers, mqttServer)
		}
	}

	return servers, nil
}

func buildNats(config conf.ServerConf, requests chan update.UpdateRecordRequest) (*sink.NatsDyndnsServer, error) {
	log.Info().Msg("Building NATS notifier")
	js, err := sink.Connect(config.NatsConfig)
	if err != nil {
		return nil, err
	}

	return sink.NewNatsDyndnsServer(&config.NatsConfig, js, requests)
}

func buildListeners(config conf.ServerConf, requests chan update.UpdateRecordRequest, creds aws.CredentialsProvider) ([]listener, error) {
	var listeners []listener
	var errs error

	if len(config.Brokers) > 0 {
		log.Info().Str("component", "server").Msg("Building MQTT listener(s)...")
		mqttListeners, err := buildMqtt(config, requests)
		if err != nil {
			errs = multierr.Append(errs, err)
		}
		for _, listener := range mqttListeners {
			listener := listener
			listeners = append(listeners, listener)
		}
	}

	if config.IsConfiguredForUpdates() {
		nats, err := buildNats(config, requests)
		if err != nil {
			errs = multierr.Append(errs, err)
		}
		listeners = append(listeners, nats)
	}

	if len(config.SqsQueue) > 0 {
		log.Info().Str("component", "server").Msg("Building AWS SQS listener...")
		sqs, err := buildSqs(config, requests, creds)
		if err != nil {
			errs = multierr.Append(errs, err)
		} else {
			listeners = append(listeners, sqs)
		}
	}

	if len(config.ListenAddr) > 0 {
		log.Info().Str("component", "server").Msg("Building HTTP listener...")
		httpServer, err := buildHttpServer(config, requests)
		if err != nil {
			errs = multierr.Append(errs, err)
		} else {
			listeners = append(listeners, httpServer)
		}
	}

	return listeners, errs
}

func buildAwsCredentialsProvider(config *conf.ServerConf) (aws.CredentialsProvider, error) {
	// TODO
	return nil, nil
}

func buildHttpServer(conf conf.ServerConf, req chan update.UpdateRecordRequest) (*http.HttpServer, error) {
	http, err := http.New(conf.ListenAddr, req)
	if err != nil {
		return nil, err
	}

	return http, nil
}
