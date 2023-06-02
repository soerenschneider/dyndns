package main

import (
	"flag"
	"fmt"
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
	"os"
	"os/user"
	"path"
	"strings"
)

var configPathPreferences = []string{
	"/etc/dyndns/client.json",
	"~/.dyndns/config.json",
}

func main() {
	metrics.Version.WithLabelValues(internal.BuildVersion, internal.CommitHash).SetToCurrentTime()

	configPath := flag.String("config", "", "Path to the config file")
	once := flag.Bool("once", false, "Do not run as a daemon")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("%s (commit: %s)", internal.BuildVersion, internal.CommitHash)
		os.Exit(0)
	}

	util.InitLogging()
	if *configPath == "" {
		*configPath = getDefaultConfigFileOrEmpty()
	}
	conf, err := conf.ReadClientConfig(*configPath)
	if err != nil {
		log.Fatal().Msgf("couldn't read config file: %v", err)
	}
	if err := env.Parse(conf); err != nil {
		log.Fatal().Msgf("%+v\n", err)
	}

	// supply once flag value
	conf.Once = *once
	conf.Print()
	RunClient(conf)
}

func getDefaultConfigFileOrEmpty() string {
	homeDir := getUserHomeDirectory()
	for _, configPath := range configPathPreferences {
		if homeDir != "" {
			if strings.HasPrefix(configPath, "~/") {
				configPath = path.Join(homeDir, configPath[2:])
			} else if strings.HasPrefix(configPath, "$HOME/") {
				configPath = path.Join(homeDir, configPath[6:])
			}
		}

		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}

func getUserHomeDirectory() string {
	usr, err := user.Current()
	if err != nil || usr == nil {
		log.Warn().Msg("Could not find user home directory")
		return ""
	}
	dir := usr.HomeDir
	return dir
}

func RunClient(config *conf.ClientConf) {
	metrics.ProcessStartTime.SetToCurrentTime()

	if nil == config {
		log.Fatal().Msg("Supplied nil config")
	}

	err := conf.ValidateConfig(config)
	if err != nil {
		log.Fatal().Msgf("Verification of config failed: %v", err)
	}
	keypair := getKeypair(config.KeyPairPath)

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

	resolver, err := buildResolver(config)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build ip resolver")
	}

	dispatchers := map[string]events.EventDispatch{}
	for _, broker := range config.Brokers {
		dispatcher, err := mqtt.NewMqttClient(broker, config.ClientId, fmt.Sprintf("dyndns/%s", config.Host), config.TlsConfig())
		if err != nil {
			log.Error().Msgf("Could not build mqtt dispatcher: %v", err)
		} else {
			dispatchers[broker] = dispatcher
		}
	}

	if len(dispatchers) == 0 {
		log.Fatal().Msg("not a single dispatcher built, exiting")
	}

	reconciler, err := client.NewReconciler(dispatchers)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build reconciler")
	}
	client, err := client.NewClient(resolver, keypair, reconciler, notificationImpl)
	if err != nil {
		log.Fatal().Msgf("could not build client: %v", err)
	}

	go reconciler.Run()
	if config.Once {
		_, err := client.ResolveSingle()
		if err != nil {
			log.Info().Msgf("Error while resolving: %v", err)
			os.Exit(1)
		}
	} else {
		go metrics.StartMetricsServer(config.MetricsListener)
		client.Run()
	}
}

func buildResolver(conf *conf.ClientConf) (resolvers.IpResolver, error) {
	if conf.InterfaceConfig != nil {
		log.Info().Msgf("Building new resolver for interface %s", conf.NetworkInterface)
		return resolvers.NewInterfaceResolver(conf.NetworkInterface, conf.Host)
	}

	log.Info().Msgf("Building HTTP resolver")
	return resolvers.NewHttpResolver(conf.Host, conf.PreferredUrls, conf.FallbackUrls, conf.AddrFamilies)
}

func getKeypair(path string) verification.SignatureKeypair {
	log.Info().Msgf("Trying to read keypair from file %s", path)
	keypair, err := verification.FromFile(path)
	if err != nil {
		log.Info().Msgf("Creating new keypair, as I couldn't read keypair: %v", err)
		keypair, err = verification.NewKeyPair()
		if err != nil {
			log.Fatal().Msgf("Can not create keypair: %v", err)
		}

		err = verification.ToFile(path, keypair)
		if err != nil {
			log.Fatal().Msgf("Could not save keypair: %v", err)
		}
	}

	return keypair
}
