package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"sync"
	"time"
)

const publishWaitTimeout = 5 * time.Minute

type MqttBus struct {
	client            mqtt.Client
	notificationTopic string
}

var mutex sync.Mutex

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Info().Msgf("Connection lost: %v", err)
	metrics.MqttConnectionsLostTotal.Inc()
	mutex.Lock()
	defer mutex.Unlock()
	metrics.MqttBrokersConnectedTotal.Sub(1)
}

func NewMqttDispatch(brokers []string, clientId, notificationTopic string, tlsConfig *tls.Config) (*MqttBus, error) {
	opts := mqtt.NewClientOptions()
	for _, broker := range brokers {
		opts.AddBroker(broker)
	}
	opts.SetClientID(clientId)

	opts.OnConnect = func(client mqtt.Client) {
		log.Info().Msgf("Connected to brokers %v", opts.Servers)
		mutex.Lock()
		metrics.MqttBrokersConnectedTotal.Add(1)
		mutex.Unlock()
	}

	opts.OnConnectionLost = connectLostHandler
	opts.AutoReconnect = true

	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}

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
		client.Subscribe(notificationTopic, 1, handler)
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
