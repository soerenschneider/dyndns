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
	log.Warn().Err(err).Str("component", "mqtt").Any("brokers", opts.Servers()).Msg("Connection lost")
	metrics.MqttConnectionsLostTotal.Inc()
	mutex.Lock()
	defer mutex.Unlock()
	metrics.MqttBrokersConnectedTotal.Sub(1)
}

func onReconnectHandler(client mqtt.Client, opts *mqtt.ClientOptions) {
	mutex.Lock()
	metrics.MqttReconnectionsTotal.Inc()
	mutex.Unlock()
	log.Info().Str("component", "mqtt").Any("brokers", opts.Servers).Msg("Reconnecting")
}

func onConnectAttemptHandler(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
	log.Info().Str("component", "mqtt").Str("broker", broker.Host).Msg("Trying connecting to broker")
	return tlsCfg
}

var onConnectHandler = func(c mqtt.Client) {
	opts := c.OptionsReader()
	log.Info().Str("component", "mqtt").Any("brokers", opts.Servers()).Msg("Successfully connected")
	mutex.Lock()
	metrics.MqttBrokersConnectedTotal.Add(1)
	mutex.Unlock()
}
