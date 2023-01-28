//go:build server

package mqtt

import (
	"crypto/tls"
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"time"
)

type MqttBus struct {
	client            mqtt.Client
	notificationTopic string
	requests          chan common.Envelope
}

func (s *MqttBus) onMessage(client mqtt.Client, msg mqtt.Message) {
	opts := client.OptionsReader()
	log.Info().Msgf("Picked up message from broker %s", opts.Servers()[0])
	var env common.Envelope
	err := json.Unmarshal(msg.Payload(), &env)
	if err != nil {
		metrics.MessageParsingFailed.Inc()
		log.Info().Msgf("Can't parse message: %v", err)
		return
	}

	s.requests <- env
}

func NewMqttServer(broker string, clientId, notificationTopic string, tlsConfig *tls.Config, reqChan chan common.Envelope) (*MqttBus, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)

	opts.SetClientID(clientId)
	opts.OnConnectionLost = connectLostHandler
	opts.AutoReconnect = true

	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}

	bus := &MqttBus{
		notificationTopic: notificationTopic,
		requests:          reqChan,
	}

	opts.OnConnect = func(client mqtt.Client) {
		log.Info().Msgf("Connected to brokers %v", opts.Servers)
		token := client.Subscribe(notificationTopic, 1, bus.onMessage)
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

	bus.client = client
	return bus, nil
}

func (d *MqttBus) Disconnect() {
	log.Info().Msg("Disconnecting from mqtt broker")
	d.client.Disconnect(5000)
}
