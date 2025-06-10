//go:build client

package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
)

const publishWaitTimeout = 10 * time.Second

type MqttClientBus struct {
	client            mqtt.Client
	notificationTopic string
}

func NewMqttClient(broker string, clientId, notificationTopic string, tlsConfig *tls.Config) (*MqttClientBus, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientId)
	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}

	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(60 * time.Second)
	opts.SetConnectRetry(true)
	opts.SetClientID(clientId)

	opts.OnConnectionLost = connectLostHandler
	opts.OnConnectAttempt = onConnectAttemptHandler
	opts.OnConnect = onConnectHandler
	opts.OnReconnecting = onReconnectHandler

	client := mqtt.NewClient(opts)
	token := client.Connect()
	finishedWithinTimeout := token.WaitTimeout(10 * time.Second)
	if token.Error() != nil || !finishedWithinTimeout {
		log.Error().Err(token.Error()).Str("component", "mqtt").Str("broker", broker).Msg("Connection to broker failed, continuing in background")
	}

	return &MqttClientBus{
		client:            client,
		notificationTopic: notificationTopic,
	}, nil
}

func (d *MqttClientBus) Notify(msg *common.UpdateRecordRequest) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal envelope: %v", err)
	}
	opts := d.client.OptionsReader()
	log.Debug().Msgf("Sending %v to %v", string(payload), opts.Servers())

	token := d.client.Publish(d.notificationTopic, 1, true, payload)
	ok := token.WaitTimeout(publishWaitTimeout)
	if !ok {
		return errors.New("received timeout when trying to publish the message")
	}
	log.Debug().Str("component", "mqtt").Any("brokers", opts.Servers()).Msg("Dispatched message")

	return nil
}
