//go:build client

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
	"time"
)

const publishWaitTimeout = 2 * time.Minute

type MqttClientBus struct {
	client            mqtt.Client
	notificationTopic string
}

var onConnectHandler = func(c mqtt.Client) {
	opts := c.OptionsReader()
	log.Info().Msgf("Connected to brokers %v", opts.Servers())
	mutex.Lock()
	metrics.MqttBrokersConnectedTotal.Add(1)
	mutex.Unlock()
}

func NewMqttClient(broker string, clientId, notificationTopic string, tlsConfig *tls.Config) (*MqttClientBus, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientId)

	opts.OnConnectionLost = connectLostHandler
	opts.OnConnect = onConnectHandler
	opts.AutoReconnect = true

	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.WaitTimeout(60*time.Second) && token.Error() != nil {
		return nil, fmt.Errorf("could not connect to %s: %v", broker, token.Error())
	}

	return &MqttClientBus{
		client:            client,
		notificationTopic: notificationTopic,
	}, nil
}

func (d *MqttClientBus) Notify(msg *common.Envelope) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal envelope: %v", err)
	}

	token := d.client.Publish(d.notificationTopic, 1, true, payload)
	ok := token.WaitTimeout(publishWaitTimeout)
	if !ok {
		return errors.New("received timeout when trying to publish the message")
	}

	return nil
}
