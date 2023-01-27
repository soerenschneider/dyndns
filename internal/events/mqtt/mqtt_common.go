package mqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"sync"
)

var mutex sync.Mutex

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	opts := client.OptionsReader()
	log.Info().Msgf("Connection lost from %v: %v", opts.Servers(), err)
	metrics.MqttConnectionsLostTotal.Inc()
	mutex.Lock()
	defer mutex.Unlock()
	metrics.MqttBrokersConnectedTotal.Sub(1)
}
