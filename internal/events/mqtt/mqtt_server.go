//go:build server

package mqtt

import (
	"crypto/tls"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"time"
)

type MqttBus struct {
	client            mqtt.Client
	notificationTopic string
}

func NewMqttServer(brokers []string, clientId, notificationTopic string, tlsConfig *tls.Config, handler func(client mqtt.Client, msg mqtt.Message)) (*MqttBus, error) {
	opts := mqtt.NewClientOptions()
	for _, broker := range brokers {
		opts.AddBroker(broker)
	}

	opts.SetClientID(clientId)
	opts.OnConnectionLost = connectLostHandler
	opts.AutoReconnect = true

	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}

	opts.OnConnect = func(client mqtt.Client) {
		log.Info().Msgf("Connected to brokers %v", opts.Servers)
		token := client.Subscribe(notificationTopic, 1, handler)
		if !token.WaitTimeout(60 * time.Second) {
			log.Error().Msgf("Could not re-subscribe to %s", notificationTopic)
			return
		}
		log.Info().Msgf("Subscribed to topic %s", notificationTopic)
		mutex.Lock()
		defer mutex.Unlock()
		metrics.MqttBrokersConnectedTotal.Add(1)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal().Msgf("Connection to broker failed: %v", token.Error())
	}

	return &MqttBus{
		client:            client,
		notificationTopic: notificationTopic,
	}, nil
}

func (d *MqttBus) Disconnect() {
	log.Info().Msg("Disconnecting from mqtt broker")
	d.client.Disconnect(5000)
}
