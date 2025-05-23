//go:build app

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
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal"
	"github.com/soerenschneider/dyndns/internal/common"
	conf "github.com/soerenschneider/dyndns/internal/conf"
	"github.com/soerenschneider/dyndns/internal/events/http"
	"github.com/soerenschneider/dyndns/internal/events/mqtt"
	sink "github.com/soerenschneider/dyndns/internal/events/nats"
	client "github.com/soerenschneider/dyndns/internal/events/sqs"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/notification"
	"github.com/soerenschneider/dyndns/internal/server"
	"github.com/soerenschneider/dyndns/internal/server/dns"
	"github.com/soerenschneider/dyndns/internal/server/vault"
	"github.com/soerenschneider/dyndns/internal/util"
	"go.uber.org/multierr"
)

const defaultConfigPath = "/etc/dyndns/config.json"
const notificationTopic = "dyndns/+"

func dieOnError(err error, msg string) {
	if err != nil {
		log.Fatal().Err(err).Msg(msg)
	}
}

type Listener interface {
	Listen(ctx context.Context, wg *sync.WaitGroup) error
}

func main() {
	configPath := flag.String("config", defaultConfigPath, "Path to the config file")
	version := flag.Bool("version", false, "Print version and exit")
	debug := flag.Bool("debug", false, "Print debug logs")
	flag.Parse()

	if *version {
		fmt.Printf("%s (commit: %s) go%s\n", internal.BuildVersion, internal.CommitHash, internal.GoVersion)
		os.Exit(0)
	}

	util.InitLogging(*debug)

	metrics.Version.WithLabelValues(internal.BuildVersion, internal.CommitHash, internal.GoVersion).Set(1)
	metrics.ProcessStartTime.SetToCurrentTime()

	config, err := conf.ReadServerConfig(*configPath)
	if err != nil {
		if *configPath != defaultConfigPath {
			dieOnError(err, "couldn't read config file")
		}
		config = conf.GetDefaultServerConfig()
	}

	err = conf.ParseEnvVariables(config)
	dieOnError(err, "could not parse env variables")

	err = conf.ValidateConfig(config)
	dieOnError(err, "Config validation failed")

	RunServer(config)
}

func RunServer(config *conf.ServerConf) {
	metrics.MqttBrokersConfiguredTotal.Set(float64(len(config.Brokers)))
	conf.PrintFields(config, conf.SensitiveFields...)

	notificationImpl, err := buildNotificationImpl(*config)
	dieOnError(err, "Can't build notification impl")

	var requestsChannel = make(chan common.UpdateRecordRequest)
	ctx, cancel := context.WithCancel(context.Background())

	go metrics.StartMetricsServer(config.MetricsListener)
	go metrics.StartHeartbeat(ctx)

	// set hash of known hosts
	hash, err := conf.GetKnownHostsHash(config.KnownHosts)
	if err != nil {
		log.Warn().Err(err).Str("component", "server").Msg("could not reliably compute hash for known_hosts, alerts may trigger")
		metrics.KnownHostsHash.Set(float64(hash))
	}

	provider, err := buildAwsCredentialsProvider(config)
	dieOnError(err, "could not build credentials provider")

	listeners, err := buildListeners(*config, requestsChannel, provider)
	if err != nil {
		log.Error().Err(err).Msg("could not build all listeners")
	}
	if len(listeners) == 0 {
		log.Fatal().Err(err).Msg("no usable listener has been built")
	}

	propagator, err := dns.NewRoute53Propagator(config.HostedZoneId, provider)
	dieOnError(err, "Could not build dns propagation implementation")

	dyndnsServer, err := server.NewServer(*config, propagator, requestsChannel, notificationImpl)
	dieOnError(err, "could not build dyndns server")

	log.Info().Str("component", "server").Msg("Ready, listening for incoming requests")
	go dyndnsServer.Listen()

	wg := &sync.WaitGroup{}
	for _, listener := range listeners {
		go func() {
			err := listener.Listen(ctx, wg)
			if err != nil {
				log.Fatal().Err(err).Str("component", "server").Msg("could not start listener")
			}
		}()
	}

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
	<-term
	log.Info().Str("component", "server").Msg("Caught signal, cancelling context")
	cancel()
	wg.Wait()
	close(requestsChannel)
}

func buildSqs(config conf.ServerConf, requests chan common.UpdateRecordRequest, credProvider credentials.Provider) (*client.SqsListener, error) {
	return client.NewSqsConsumer(config.SqsConfig, credProvider, requests)
}

func buildMqtt(config conf.ServerConf, requests chan common.UpdateRecordRequest) ([]*mqtt.MqttBus, error) {
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

func buildNats(config conf.ServerConf, requests chan common.UpdateRecordRequest) (*sink.NatsDyndnsServer, error) {
	log.Info().Msg("Building NATS notifier")
	js, err := sink.Connect(config.NatsConfig)
	if err != nil {
		return nil, err
	}

	return sink.NewNatsDyndnsServer(&config.NatsConfig, js, requests)
}

func buildListeners(config conf.ServerConf, requests chan common.UpdateRecordRequest, creds credentials.Provider) ([]Listener, error) {
	var listeners []Listener
	var errs error

	if len(config.MqttConfig.Brokers) > 0 {
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

	if config.NatsConfig.IsConfiguredForUpdates() {
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

	if len(config.HttpConfig.ListenAddr) > 0 {
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

func buildAwsCredentialsProvider(config *conf.ServerConf) (credentials.Provider, error) {
	if !config.UseVaultCredentialsProvider() {
		return nil, nil
	}

	client, err := buildVaultClient(config.VaultConfig)
	if err != nil {
		return nil, err
	}
	auth, err := buildVaultAuth(config.VaultConfig)
	if err != nil {
		return nil, err
	}

	return buildAwsVaultCredentialProvider(&config.VaultConfig, client, auth)
}

func buildVaultClient(conf conf.VaultConfig) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = conf.VaultAddr
	config.Timeout = 30 * time.Second

	return api.NewClient(config)
}

func buildHttpServer(conf conf.ServerConf, req chan common.UpdateRecordRequest) (*http.HttpServer, error) {
	http, err := http.New(conf.HttpConfig.ListenAddr, req)
	if err != nil {
		return nil, err
	}

	return http, nil
}

// buildAwsVaultCredentialProvider returns the vault credentials provider, but only if it succeeds to login at vault
// otherwise the default credentials provider by AWS is used, trying to be resilient
func buildAwsVaultCredentialProvider(config *conf.VaultConfig, client *api.Client, auth vault.Auth) (credentials.Provider, error) {
	if config == nil {
		return nil, errors.New("nil config provided")
	}

	provider, err := vault.NewVaultCredentialProvider(client, auth, config)
	if err != nil {
		return nil, err
	}

	log.Info().Str("component", "server").Msg("Testing authentication against vault")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err = client.Auth().Login(ctx, auth)
	if err != nil {
		return nil, fmt.Errorf("could not authenticate against vault: %w", err)
	}

	return provider, nil
}

func buildVaultAuth(config conf.VaultConfig) (vault.Auth, error) {
	switch config.AuthStrategy {
	case conf.VaultAuthStrategyToken:
		return vault.NewTokenAuth(config.VaultToken)
	case conf.VaultAuthStrategyApprole:
		secretId := &approle.SecretID{
			FromString: config.AppRoleSecretId,
		}
		return approle.NewAppRoleAuth(config.AppRoleId, secretId)
	default:
		return nil, errors.New("can't build vault auth")
	}
}
func buildNotificationImpl(config conf.ServerConf) (notification.Notification, error) {
	if config.EmailConfig.IsConfigured() {
		err := config.EmailConfig.Validate()
		if err != nil {
			log.Fatal().Err(err).Msgf("Bad email config")
		}
		return util.NewEmailNotification(&config.EmailConfig)
	}

	if config.NatsConfig.IsConfiguredAsListener() {
		js, err := sink.Connect(config.NatsConfig)
		if err != nil {
			return nil, err
		}
		return sink.NewNatsCloudevents(&config.NatsConfig, js)
	}

	return &notification.DummyNotification{}, nil
}
