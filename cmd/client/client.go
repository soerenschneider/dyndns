package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/client"
	"github.com/soerenschneider/dyndns/client/resolvers"
	"github.com/soerenschneider/dyndns/conf"
	"github.com/soerenschneider/dyndns/internal"
	"github.com/soerenschneider/dyndns/internal/events"
	"github.com/soerenschneider/dyndns/internal/events/mqtt"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/notification"
	"github.com/soerenschneider/dyndns/internal/util"
	"github.com/soerenschneider/dyndns/internal/verification"
	"github.com/soerenschneider/dyndns/internal/verification/key_provider"
)

var (
	configPath    string
	once          bool
	cmdVersion    bool
	cmdGenKeypair bool
)

func main() {
	parseFlags()

	if cmdVersion {
		fmt.Printf("%s (commit: %s)", internal.BuildVersion, internal.CommitHash)
		os.Exit(0)
	}

	if cmdGenKeypair {
		generateKeypair()
	}

	util.InitLogging()
	if configPath == "" {
		configPath = conf.GetDefaultConfigFileOrEmpty()
	}

	config, err := conf.ReadClientConfig(configPath)
	if err != nil {
		log.Warn().Err(err).Msgf("couldn't read config file")
	}

	if err := env.Parse(config); err != nil {
		log.Fatal().Err(err).Msg("could not parse env variables")
	}

	if err := conf.ValidateConfig(config); err != nil {
		log.Fatal().Msgf("Verification of config failed: %v", err)
	}

	metrics.MqttBrokersConfiguredTotal.Set(float64(len(config.Brokers)))

	// supply once flag value
	config.Once = once

	conf.PrintFields(config, conf.SensitiveFields...)
	RunClient(config)
}

func parseFlags() {
	flag.StringVar(&configPath, "config", "", "Path to the config file")
	flag.BoolVar(&once, "once", false, "Do not run as a daemon")
	flag.BoolVar(&cmdVersion, "version", false, "Print version and exit")
	flag.BoolVar(&cmdGenKeypair, "gen-keypair", false, "Generate keypair")
	flag.Parse()
}

func dieOnError(err error, msg string) {
	if err != nil {
		log.Fatal().Err(err).Msg(msg)
	}
}

func RunClient(config *conf.ClientConf) {
	metrics.Version.WithLabelValues(internal.BuildVersion, internal.CommitHash).SetToCurrentTime()
	metrics.ProcessStartTime.SetToCurrentTime()

	provider, err := buildKeyProvider(config)
	dieOnError(err, "can not build key key_provider")

	keypair, err := getKeypair(provider)
	dieOnError(err, "can not get keypair")

	notificationImpl, err := buildNotificationImpl(config)
	dieOnError(err, "Can't build email notification")

	resolver, err := buildResolver(config)
	dieOnError(err, "could not build ip resolver")

	dispatchers := map[string]events.EventDispatch{}
	for _, broker := range config.Brokers {
		dispatchers[broker], err = mqtt.NewMqttClient(broker, config.ClientId, fmt.Sprintf("dyndns/%s", config.Host), config.TlsConfig())
		dieOnError(err, "Could not build mqtt dispatcher")
	}

	reconciler, err := client.NewReconciler(dispatchers)
	dieOnError(err, "could not build reconciler")

	client, err := client.NewClient(resolver, keypair, reconciler, notificationImpl)
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
		log.Info().Msgf("Building new resolver for interface %s", conf.NetworkInterface)
		return resolvers.NewInterfaceResolver(conf.NetworkInterface, conf.Host)
	}

	log.Info().Msgf("Building HTTP resolver")
	return resolvers.NewHttpResolver(conf.Host, conf.PreferredUrls, conf.FallbackUrls, conf.AddrFamilies)
}

func buildNotificationImpl(config *conf.ClientConf) (notification.Notification, error) {
	if config.EmailConfig != nil {
		err := config.EmailConfig.Validate()
		if err != nil {
			log.Fatal().Err(err).Msgf("Bad email config")
		}
		return util.NewEmailNotification(config.EmailConfig)
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
	log.Info().Msg("Trying to read keypair")
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

	log.Info().Msgf("Creating new keypair, as I couldn't read keypair: %v", err)
	keypair, err = verification.NewKeyPair()
	if err != nil {
		log.Fatal().Msgf("Can not create keypair: %v", err)
	}
	log.Info().Msgf("Created keypair with pubkey '%s'", base64.StdEncoding.EncodeToString(keypair.PubKey))

	jsonData, err := keypair.AsJson()
	if err != nil {
		return nil, err
	}

	err = provider.Write(jsonData)
	if err != nil {
		log.Fatal().Msgf("Could not save keypair: %v", err)
	}

	return keypair, nil
}

func generateKeypair() {
	keypair, err := verification.NewKeyPair()
	if err != nil {
		log.Fatal().Msgf("Can not create keypair: %v", err)
	}

	jsonEncoded, err := keypair.AsJson()
	if err != nil {
		log.Fatal().Err(err).Msg("could not marshall keypair")
	}
	fmt.Printf("%s\n", jsonEncoded)
	os.Exit(0)
}
