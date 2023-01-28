package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/conf"
	"github.com/soerenschneider/dyndns/internal"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/events/mqtt"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/util"
	"github.com/soerenschneider/dyndns/server"
	"github.com/soerenschneider/dyndns/server/dns"
	"github.com/soerenschneider/dyndns/server/vault"
	"os"
	"os/signal"
	"syscall"
)

const defaultConfigPath = "/etc/dyndns/config.json"
const notificationTopic = "dyndns/+"

func main() {
	metrics.Version.WithLabelValues(internal.BuildVersion, internal.CommitHash).SetToCurrentTime()
	configPath := flag.String("config", defaultConfigPath, "Path to the config file")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("%s (commit: %s)", internal.BuildVersion, internal.CommitHash)
		os.Exit(0)
	}

	util.InitLogging()
	if nil == configPath {
		log.Fatal().Msgf("No config path supplied")
	}

	RunServer(*configPath)
}

// getCredentialProvider returns the vault credentials provider, but only if it succeeds to login at vault
// otherwise the default credentials provider by AWS is used, trying to be resilient
func getCredentialProvider(config conf.VaultConfig) credentials.Provider {
	if config.Verify() != nil {
		return nil
	}

	provider, err := vault.NewVaultCredentialProvider(&config)
	if err != nil {
		log.Info().Msgf("couldn't build vault dynamic credential provider: %v", err)
		// TODO: metrics
		return nil
	}

	log.Info().Msg("Testing authentication against vault")
	err = provider.LookupToken()
	if err != nil {
		log.Error().Msgf("Could not authenticate against vault: %v", err)
		// TODO: metrics
		return nil
	}

	return provider
}

func RunServer(configPath string) {
	metrics.ProcessStartTime.SetToCurrentTime()

	conf, err := conf.ReadServerConfig(configPath)
	if err != nil {
		log.Fatal().Msgf("couldn't read config file: %v", err)
	}

	err = conf.Validate()
	if err != nil {
		log.Fatal().Msgf("Config validation failed: %v", err)
	}
	conf.Print()

	var requestsChannel = make(chan common.Envelope)
	var servers []*mqtt.MqttBus
	for _, broker := range conf.Brokers {
		mqttServer, err := mqtt.NewMqttServer(broker, conf.ClientId, notificationTopic, conf.TlsConfig(), requestsChannel)
		if err != nil {
			log.Fatal().Msgf("Could not build mqtt dispatcher: %v", err)
		}
		servers = append(servers, mqttServer)
	}

	go metrics.StartMetricsServer(conf.MetricsListener)

	provider := getCredentialProvider(conf.VaultConfig)
	propagator, err := dns.NewRoute53Propagator(conf.HostedZoneId, provider)
	if err != nil {
		log.Fatal().Msgf("Could not build dns propagation implementation: %v", err)
	}

	dyndnsServer, err := server.NewServer(*conf, propagator, requestsChannel)
	if err != nil {
		log.Fatal().Msgf("Could not build server: %v", err)
	}
	go dyndnsServer.Listen()

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
	<-term
	log.Info().Msg("Caught signal")
	for index := range servers {
		mqttServer := servers[index]
		mqttServer.Disconnect()
	}

	close(requestsChannel)
}
