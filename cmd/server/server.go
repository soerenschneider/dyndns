package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/conf"
	"github.com/soerenschneider/dyndns/internal"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/events/mqtt"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/notification"
	"github.com/soerenschneider/dyndns/internal/util"
	"github.com/soerenschneider/dyndns/server"
	"github.com/soerenschneider/dyndns/server/dns"
	vaultDyndns "github.com/soerenschneider/dyndns/server/vault"
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

func RunServer(configPath string) {
	metrics.ProcessStartTime.SetToCurrentTime()

	config, err := conf.ReadServerConfig(configPath)
	if err != nil {
		log.Fatal().Msgf("couldn't read config file: %v", err)
	}

	err = conf.ValidateConfig(config)
	if err != nil {
		log.Fatal().Msgf("Config validation failed: %v", err)
	}
	metrics.MqttBrokersConfiguredTotal.Set(float64(len(config.Brokers)))
	conf.PrintFields(config, conf.SensitiveFields...)

	var notificationImpl notification.Notification
	if config.EmailConfig != nil {
		err := config.EmailConfig.Validate()
		if err != nil {
			log.Fatal().Err(err).Msgf("Bad email config")
		}
		notificationImpl, err = util.NewEmailNotification(config.EmailConfig)
		if err != nil {
			log.Fatal().Err(err).Msgf("Can't build email notification")
		}
	}

	var requestsChannel = make(chan common.Envelope)
	var servers []*mqtt.MqttBus
	for _, broker := range config.Brokers {
		mqttServer, err := mqtt.NewMqttServer(broker, config.ClientId, notificationTopic, config.TlsConfig(), requestsChannel)
		if err != nil {
			log.Fatal().Msgf("Could not build mqtt dispatcher: %v", err)
		}
		servers = append(servers, mqttServer)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go metrics.StartMetricsServer(config.MetricsListener)
	go metrics.StartHeartbeat(ctx)

	// set hash of known hosts
	hash, err := conf.GetKnownHostsHash(config.KnownHosts)
	if err != nil {
		log.Warn().Err(err).Msg("could not reliably compute hash for known_hosts, alerts may trigger")
		metrics.KnownHostsHash.Set(float64(hash))
	}

	var provider credentials.Provider
	if config.UseVaultCredentialsProvider() {
		client, err := buildVaultClient(config.VaultConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("could not build vault client")
		}
		auth, err := buildVaultAuth(config.VaultConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("could not build auth")
		}
		provider, err = buildCredentialProvider(config.VaultConfig, client, auth)
		if err != nil {
			log.Fatal().Err(err).Msg("could not build credentials provider")
		}
	}

	propagator, err := dns.NewRoute53Propagator(config.HostedZoneId, provider)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not build dns propagation implementation")
	}

	dyndnsServer, err := server.NewServer(*config, propagator, requestsChannel, notificationImpl)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build dyndns server")
	}
	go dyndnsServer.Listen()

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
	<-term
	log.Info().Msg("Caught signal, cancelling context")
	cancel()
	for index := range servers {
		mqttServer := servers[index]
		mqttServer.Disconnect()
	}

	close(requestsChannel)
}

func buildVaultClient(conf *conf.VaultConfig) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = conf.VaultAddr
	config.Timeout = 30 * time.Second

	return api.NewClient(config)
}

// buildCredentialProvider returns the vault credentials provider, but only if it succeeds to login at vault
// otherwise the default credentials provider by AWS is used, trying to be resilient
func buildCredentialProvider(config *conf.VaultConfig, client *api.Client, auth vaultDyndns.Auth) (credentials.Provider, error) {
	if config == nil {
		return nil, errors.New("nil config provided")
	}

	provider, err := vaultDyndns.NewVaultCredentialProvider(client, auth, config)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Testing authentication against vault")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err = client.Auth().Login(ctx, auth)
	if err != nil {
		return nil, fmt.Errorf("could not authenticate against vault: %w", err)
	}

	return provider, nil
}

func buildVaultAuth(config *conf.VaultConfig) (vaultDyndns.Auth, error) {
	switch config.AuthStrategy {
	case conf.VaultAuthStrategyToken:
		return vaultDyndns.NewTokenAuth(config.VaultToken)
	case conf.VaultAuthStrategyApprole:
		secretId := &approle.SecretID{
			FromString: config.AppRoleSecretId,
		}
		return approle.NewAppRoleAuth(config.AppRoleId, secretId)
	default:
		return nil, errors.New("can't build vault auth")
	}
}
