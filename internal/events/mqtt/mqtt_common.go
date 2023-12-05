package mqtt

import (
	"crypto/tls"
	"net/url"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/metrics"
)

var mutex sync.Mutex

func connectLostHandler(client mqtt.Client, err error) {
	opts := client.OptionsReader()
	log.Info().Msgf("Connection lost from %v: %v", opts.Servers(), err)
	metrics.MqttConnectionsLostTotal.Inc()
	mutex.Lock()
	defer mutex.Unlock()
	metrics.MqttBrokersConnectedTotal.Sub(1)
}

func onReconnectHandler(client mqtt.Client, opts *mqtt.ClientOptions) {
	mutex.Lock()
	metrics.MqttReconnectionsTotal.Inc()
	mutex.Unlock()
	log.Info().Msgf("Reconnecting to %s", opts.Servers)
}

func onConnectAttemptHandler(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
	log.Info().Msgf("Attempting to connect to broker %s", broker.Host)
	return tlsCfg
}

var onConnectHandler = func(c mqtt.Client) {
	opts := c.OptionsReader()
	log.Info().Msgf("Connected to broker(s) %v", opts.Servers())
	mutex.Lock()
	metrics.MqttBrokersConnectedTotal.Add(1)
	mutex.Unlock()
}
