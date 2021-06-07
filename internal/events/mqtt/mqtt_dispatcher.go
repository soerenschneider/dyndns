package mqtt

import (
	"dyndns/internal/common"
	"dyndns/internal/metrics"
	"encoding/json"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"time"
)

const publishWaitTimeout = 5 * time.Minute

type MqttBus struct {
	client            mqtt.Client
	notificationTopic string
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Info().Msg("Successfully connected to broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Info().Msgf("Connection lost: %v", err)
	metrics.MqttConnectionsLostTotal.Inc()
}

func NewMqttDispatch(broker, clientId, notificationTopic string) (*MqttBus, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientId)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.AutoReconnect = true

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.WaitTimeout(60*time.Second) && token.Error() != nil {
		log.Fatal().Msgf("Connection to broker failed: %v", token.Error())
	}

	return &MqttBus{
		client:            client,
		notificationTopic: notificationTopic,
	}, nil
}

func NewMqttServer(broker, clientId, notificationTopic string, handler func(client mqtt.Client, msg mqtt.Message)) (*MqttBus, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientId)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.AutoReconnect = true

	opts.OnConnect = func(client mqtt.Client) {
		log.Info().Msgf("Connected to brokers %v", opts.Servers)
		client.Subscribe(notificationTopic, 1, handler)
		log.Info().Msgf("Subscribed to topic %s", notificationTopic)
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

func (d *MqttBus) Notify(msg common.Envelope) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal envelope: %v", err)
	}

	token := d.client.Publish(d.notificationTopic, 1, true, payload)
	ok := token.WaitTimeout(publishWaitTimeout)
	if ok {
		return nil
	}

	return errors.New("received timeout when trying to publish the message")
}
