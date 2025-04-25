//go:build server

package mqtt

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/metrics"
)

type MqttBus struct {
	client            mqtt.Client
	notificationTopic string
	broker            string

	requests chan common.UpdateRecordRequest
}

func NewMqttServer(broker string, clientId, notificationTopic string, tlsConfig *tls.Config, reqChan chan common.UpdateRecordRequest) (*MqttBus, error) {
	bus := &MqttBus{
		notificationTopic: notificationTopic,
		requests:          reqChan,
		broker:            broker,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}

	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(60 * time.Second)
	opts.SetConnectRetry(true)
	opts.SetClientID(clientId)

	opts.OnConnectionLost = connectLostHandler
	opts.OnConnectAttempt = onConnectAttemptHandler
	opts.OnConnect = bus.onConnect
	opts.OnReconnecting = onReconnectHandler

	bus.client = mqtt.NewClient(opts)

	return bus, nil
}

func (s *MqttBus) Listen(ctx context.Context, wg *sync.WaitGroup) error {
	token := s.client.Connect()
	finishedWithinTimeout := token.WaitTimeout(10 * time.Second)
	if token.Error() != nil || !finishedWithinTimeout {
		log.Error().Err(token.Error()).Str("broker", s.broker).Msg("Connection to broker failed, continuing in background")
	}

	return nil
}

func (s *MqttBus) Disconnect() {
	log.Info().Str("component", "mqtt").Str("broker", s.broker).Msg("Disconnecting from mqtt broker")
	s.client.Disconnect(5000)
}

func (s *MqttBus) onMessage(_ mqtt.Client, msg mqtt.Message) {
	log.Info().Str("component", "mqtt").Str("broker", s.broker).Msg("Picked up message")
	var env common.UpdateRecordRequest
	err := json.Unmarshal(msg.Payload(), &env)
	if err != nil {
		metrics.MessageParsingFailed.Inc()
		log.Warn().Err(err).Str("component", "mqtt").Str("broker", s.broker).Msg("Can't parse message")
		return
	}

	s.requests <- env
}

func (s *MqttBus) onConnect(client mqtt.Client) {
	log.Info().Str("component", "mqtt").Str("broker", s.broker).Msgf("Connected to broker")
	token := client.Subscribe(s.notificationTopic, 1, s.onMessage)
	if !token.WaitTimeout(60 * time.Second) {
		log.Error().Msgf("Could not re-subscribe to %s", s.notificationTopic)
		return
	}

	log.Info().Str("component", "mqtt").Str("broker", s.broker).Str("topic", s.notificationTopic).Msg("Subscribed to topic")
	mutex.Lock()
	defer mutex.Unlock()
	metrics.MqttBrokersConnectedTotal.Add(1)
}
