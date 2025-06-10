package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal"
	"github.com/soerenschneider/dyndns/internal/client"
	"github.com/soerenschneider/dyndns/internal/client/dispatchers"
	"github.com/soerenschneider/dyndns/internal/client/resolvers"
	"github.com/soerenschneider/dyndns/internal/conf"
	"github.com/soerenschneider/dyndns/internal/events/mqtt"
	sink "github.com/soerenschneider/dyndns/internal/events/nats"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/notification"
	"github.com/soerenschneider/dyndns/internal/util"
	"github.com/soerenschneider/dyndns/internal/verification"
	"github.com/soerenschneider/dyndns/internal/verification/key_provider"
	"go.uber.org/multierr"
)

var (
	configPath      string
	once            bool
	forceSendUpdate bool
	debug           bool
	cmdVersion      bool
	cmdGenKeypair   bool
)

func main() {
	parseFlags()

	if cmdVersion {
		fmt.Printf("%s (commit: %s) go%s\n", internal.BuildVersion, internal.CommitHash, internal.GoVersion)
		os.Exit(0)
	}

	if cmdGenKeypair {
		generateKeypair()
	}

	util.InitLogging(debug)
	if configPath == "" {
		configPath = conf.GetDefaultConfigFileOrEmpty()
	}

	config, err := conf.ReadClientConfig(configPath)
	dieOnError(err, "couldn't read config file")
	fmt.Println(config)

	err = conf.ParseClientConfEnv(config)
	dieOnError(err, "could not parse env variables")

	err = conf.ValidateConfig(config)
	dieOnError(err, "Verification of config failed")

	metrics.MqttBrokersConfiguredTotal.Set(float64(len(config.Brokers)))

	// supply once flag value
	config.Once = once

	conf.PrintFields(config, conf.SensitiveFields...)
	RunClient(config)
}

func parseFlags() {
	flag.StringVar(&configPath, "config", "", "Path to the config file")
	flag.BoolVar(&once, "once", false, "Do not run as a daemon")
	flag.BoolVar(&forceSendUpdate, "force", false, "Force sending an update request at start")
	flag.BoolVar(&cmdVersion, "version", false, "Print version and exit")
	flag.BoolVar(&cmdGenKeypair, "gen-keypair", false, "Generate keypair")
	flag.BoolVar(&debug, "debug", false, "Print debug logs")
	flag.Parse()
}

func dieOnError(err error, msg string) {
	if err != nil {
		log.Fatal().Err(err).Msg(msg)
	}
}

// nolint cyclop
func buildNotifiers(config *conf.ClientConf) (map[string]client.EventDispatch, error) {
	disp := map[string]client.EventDispatch{}

	var errs error
	if len(config.Brokers) > 0 {
		log.Info().Str("component", "client").Msg("Building MQTT notifier(s)")
		for _, broker := range config.Brokers {
			dispatcher, err := mqtt.NewMqttClient(broker, config.ClientId, fmt.Sprintf("dyndns/%s", config.Host), config.TlsConfig())
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				disp[broker] = dispatcher
			}
		}
	}

	if config.IsConfiguredForUpdates() {
		log.Info().Msg("Building NATS notifier")
		js, err := sink.Connect(config.NatsConfig)
		if err != nil {
			errs = multierr.Append(errs, err)
		} else {
			dispatcher, err := sink.NewNatsDyndnsClient(&config.NatsConfig, js)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				disp[config.Url] = dispatcher
			}
		}
	}

	if len(config.HttpDispatcherConf) > 0 {
		log.Info().Str("component", "client").Msg("Building HTTP notifier")
		for _, dispatcher := range config.HttpDispatcherConf {
			httpDispatcher, err := dispatchers.NewHttpDispatcher(dispatcher.Url)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				disp[dispatcher.Url] = httpDispatcher
			}
		}
	}

	if len(config.SqsQueue) > 0 {
		log.Info().Str("component", "client").Msg("Building AWS SQS notifier")
		sqs, err := dispatchers.NewSqsDispatcher(config.SqsConfig, nil)
		if err != nil {
			errs = multierr.Append(errs, err)
		} else {
			disp["sqs"] = sqs
		}
	}

	return disp, errs
}

func RunClient(config *conf.ClientConf) {
	metrics.Version.WithLabelValues(internal.BuildVersion, internal.CommitHash, internal.GoVersion).Set(1)
	metrics.ProcessStartTime.SetToCurrentTime()

	provider, err := buildKeyProvider(config)
	dieOnError(err, "can not build key key_provider")

	keypair, err := getKeypair(provider)
	dieOnError(err, "can not get keypair")

	notificationImpl, err := buildNotificationImpl(config)
	dieOnError(err, "Can't build email notification")

	resolver, err := buildResolver(config)
	dieOnError(err, "could not build ip resolver")

	dispatchers, err := buildNotifiers(config)
	if len(dispatchers) == 0 {
		log.Fatal().Str("component", "client").Err(err).Msg("no dispatchers built")
	}
	if err != nil {
		log.Error().Str("component", "client").Err(err).Msg("could not build all dispatchers")
	}

	reconciler, err := client.NewReconciler(dispatchers, true)
	dieOnError(err, "could not build reconciler")

	opts := []client.Opts{
		client.WithInterval(15 * time.Second),
	}

	if forceSendUpdate {
		opts = append(opts, client.WithForceSendUpdate())
	}

	client, err := client.NewClient(resolver, keypair, reconciler, notificationImpl, opts...)
	dieOnError(err, "could not build client")

	go reconciler.Run()
	if config.Once {
		_, err := client.Resolve(nil)
		dieOnError(err, "error resolving ip")
	} else {
		go metrics.StartMetricsServer(config.MetricsListener)
		client.Run()
	}
}

func buildResolver(conf *conf.ClientConf) (resolvers.IpResolver, error) {
	if len(conf.NetworkInterface) > 0 {
		log.Info().Str("component", "client").Msgf("Building new resolver for interface %s", conf.NetworkInterface)
		return resolvers.NewInterfaceResolver(conf.NetworkInterface, conf.Host)
	}

	log.Info().Str("component", "client").Msg("Building HTTP resolver")
	return resolvers.NewHttpResolver(conf.Host, conf.PreferredUrls, conf.FallbackUrls, conf.AddrFamilies)
}

func buildNotificationImpl(config *conf.ClientConf) (notification.Notification, error) {
	if config.IsConfigured() {
		err := config.Validate()
		dieOnError(err, "Bad email config")
		return util.NewEmailNotification(&config.EmailConfig)
	}

	if config.SupportsCloudeventsDispatch() {
		jetstream, err := sink.Connect(config.NatsConfig)
		dieOnError(err, "could not build nats jetstream")
		return sink.NewNatsCloudevents(&config.NatsConfig, jetstream)
	}

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
