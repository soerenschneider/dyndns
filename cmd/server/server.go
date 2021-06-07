package main

import (
	"dyndns/conf"
	"dyndns/internal"
	"dyndns/internal/common"
	"dyndns/internal/events/mqtt"
	"dyndns/internal/metrics"
	"dyndns/server"
	"dyndns/server/dns"
	"dyndns/server/vault"
	"encoding/json"
	"flag"
	"github.com/aws/aws-sdk-go/aws/credentials"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const defaultConfigPath = "/etc/dyndns/config.json"
const notificationTopic = "dyndns/+"

var requestsChannel = make(chan common.Envelope)

func main() {
	log.Printf("Started dyndns server version %s, commit %s, built at %s", internal.BuildVersion, internal.CommitHash, internal.BuildTime)
	configPath := flag.String("config", defaultConfigPath, "Path to the config file")
	flag.Parse()

	if nil == configPath {
		log.Fatalf("No config path supplied")
	}

	RunServer(*configPath)
}

func HandleChangeRequest(client paho.Client, msg paho.Message) {
	var env common.Envelope
	err := json.Unmarshal(msg.Payload(), &env)
	if err != nil {
		metrics.MessageParsingFailed.Inc()
		log.Printf("Can't parse message: %v", err)
		return
	}

	requestsChannel <- env
}

// getCredentialProvider returns the vault credentials provider, but only if it succeeds to login at vault
// otherwise the default credentials provider by AWS is used, trying to be resilient
func getCredentialProvider(config conf.VaultConfig) credentials.Provider {
	if config.Verify() != nil {
		return nil
	}

	provider, err := vault.NewVaultCredentialProvider(&config)
	if err != nil {
		log.Printf("couldn't build vault dynamic credential provider: %v", err)
		// TODO: metrics
		return nil
	}

	log.Println("Testing authentication against vault")
	err = provider.LookupToken()
	if err != nil {
		log.Printf("Could not authenticate against vault: %v", err)
		// TODO: metrics
		return nil
	}

	return provider
}

func RunServer(configPath string) {
	conf, err := conf.ReadServerConfig(configPath)
	if err != nil {
		log.Fatalf("couldn't read config file: %v", err)
	}

	err = conf.Validate()
	if err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}
	conf.Print()

	mqttServer, err := mqtt.NewMqttServer(conf.Broker, conf.ClientId, notificationTopic, HandleChangeRequest)
	if err != nil {
		log.Fatalf("Could not build mqtt dispatcher: %v", err)
	}

	go metrics.StartMetricsServer(conf.MetricsListener)

	provider := getCredentialProvider(conf.VaultConfig)
	propagator, err := dns.NewRoute53Propagator(conf.HostedZoneId, provider)
	if err != nil {
		log.Fatalf("Could not build dns propagation implementation: %v", err)
	}

	dyndnsServer, err := server.NewServer(*conf, propagator, requestsChannel)
	if err != nil {
		log.Fatalf("Could not build server: %v", err)
	}
	go dyndnsServer.Listen()

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
	<-term
	log.Println("Caught signal")
	mqttServer.Disconnect()

	close(requestsChannel)
}
