package main

import (
	"dyndns/client"
	"dyndns/client/resolvers"
	"dyndns/conf"
	"dyndns/internal"
	"dyndns/internal/events"
	"dyndns/internal/events/mqtt"
	"dyndns/internal/metrics"
	"dyndns/internal/verification"
	"flag"
	"fmt"
	"log"
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
	metrics.Version.WithLabelValues(internal.BuildVersion, internal.CommitHash, internal.BuildTime).SetToCurrentTime()
	log.Printf("Started dyndns client version %s, commit %s, built at %s", internal.BuildVersion, internal.CommitHash, internal.BuildTime)
	defaultConfigPath := checkDefaultConfigFiles()
	configPath := flag.String("config", defaultConfigPath, "Path to the config file")
	once := flag.Bool("once", false, "Path to the config file")
	flag.Parse()

	if nil == configPath {
		log.Fatalf("No config path supplied")
	}

	conf, err := conf.ReadClientConfig(*configPath)
	if err != nil {
		log.Fatalf("couldn't read config file: %v", err)
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
		log.Fatal("Supplied nil config")
	}

	err := conf.Validate()
	if err != nil {
		log.Fatalf("Verification of config failed: %v", err)
	}
	keypair := getKeypair(conf.KeyPairPath)

	var resolver resolvers.IpResolver
	if conf.InterfaceConfig != nil {
		log.Printf("Building new resolver for interface %s", conf.NetworkInterface)
		resolver, _ = resolvers.NewInterfaceResolver(conf.NetworkInterface, conf.Host)
	} else {
		log.Printf("Building HTTP resolver")
		resolver, _ = resolvers.NewHttpResolver(conf.Host)
	}

	var dispatcher events.EventDispatch
	dispatcher, err = mqtt.NewMqttDispatch(conf.Broker, conf.Host, fmt.Sprintf("dyndns/%s", conf.Host))
	if err != nil {
		log.Fatalf("Could not build mqtt dispatcher: %v", err)
	}

	go metrics.StartMetricsServer(conf.MetricsListener)

	client, err := client.NewClient(resolver, keypair, dispatcher)
	if err != nil {
		log.Fatalf("could not build client: %v", err)
	}

	if conf.Once {
		_, err := client.ResolveSingle()
		if err != nil {
			log.Printf("Error while resolving: %v", err)
			os.Exit(1)
		}
	} else {
		client.Run()
	}
}

func getKeypair(path string) verification.SignatureKeypair {
	log.Printf("Trying to read keypair from file %s", path)
	keypair, err := verification.FromFile(path)
	if err != nil {
		log.Printf("Creating new keypair, as I couldn't read keypair: %v", err)
		keypair, err = verification.NewKeyPair()
		if err != nil {
			log.Fatalf("Can not create keypair: %v", err)
		}

		err = verification.ToFile(path, keypair)
		if err != nil {
			log.Fatalf("Could not save keypair: %v", err)
		}
	}

	return keypair
}
