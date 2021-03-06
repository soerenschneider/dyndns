package main

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/client"
	"github.com/soerenschneider/dyndns/client/resolvers"
	"github.com/soerenschneider/dyndns/conf"
	"github.com/soerenschneider/dyndns/internal"
	"github.com/soerenschneider/dyndns/internal/events"
	"github.com/soerenschneider/dyndns/internal/events/mqtt"
	"github.com/soerenschneider/dyndns/internal/metrics"
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
	defaultConfigPath := checkDefaultConfigFiles()
	configPath := flag.String("config", defaultConfigPath, "Path to the config file")
	once := flag.Bool("once", false, "Do not run as a daemon")
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

	conf, err := conf.ReadClientConfig(*configPath)
	if err != nil {
		log.Fatal().Msgf("couldn't read config file: %v", err)
	}
	// supply once flag value
	conf.Once = *once
	conf.Print()
	RunClient(conf)
}

func checkDefaultConfigFiles() string {
	for _, configPath := range configPathPreferences {
		if strings.HasPrefix(configPath, "~/") {
			configPath = path.Join(getUserHomeDirectory(), configPath[2:])
		} else if strings.HasPrefix(configPath, "$HOME/") {
			configPath = path.Join(getUserHomeDirectory(), configPath[6:])
		}

		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return configPathPreferences[0]
}

func getUserHomeDirectory() string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	return dir
}

func RunClient(conf *conf.ClientConf) {
	if nil == conf {
		log.Fatal().Msg("Supplied nil config")
	}

	err := conf.Validate()
	if err != nil {
		log.Fatal().Msgf("Verification of config failed: %v", err)
	}
	keypair := getKeypair(conf.KeyPairPath)

	var resolver resolvers.IpResolver
	if conf.InterfaceConfig != nil {
		log.Info().Msgf("Building new resolver for interface %s", conf.NetworkInterface)
		resolver, _ = resolvers.NewInterfaceResolver(conf.NetworkInterface, conf.Host)
	} else {
		log.Info().Msgf("Building HTTP resolver")
		resolver, _ = resolvers.NewHttpResolver(conf.Host)
	}

	var dispatchers []events.EventDispatch
	for _, broker := range conf.Brokers {
		dispatcher, err := mqtt.NewMqttDispatch(broker, conf.Host, fmt.Sprintf("dyndns/%s", conf.Host))
		if err != nil {
			log.Fatal().Msgf("Could not build mqtt dispatcher: %v", err)
		}
		dispatchers = append(dispatchers, dispatcher)
	}

	client, err := client.NewClient(resolver, keypair, dispatchers)
	if err != nil {
		log.Fatal().Msgf("could not build client: %v", err)
	}

	if conf.Once {
		_, err := client.ResolveSingle()
		if err != nil {
			log.Info().Msgf("Error while resolving: %v", err)
			os.Exit(1)
		}
	} else {
		go metrics.StartMetricsServer(conf.MetricsListener)
		client.Run()
	}
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
