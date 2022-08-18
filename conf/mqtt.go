package conf

import (
	"fmt"
	"github.com/rs/zerolog/log"
)

type MqttConfig struct {
	Brokers  []string `json:"brokers" env:"DYNDNS_BROKERS" envSeparator:";"`
	ClientId string   `json:"client_id" env:"DYNDNS_CLIENT_ID"`
}

func (conf *MqttConfig) Print() {
	log.Info().Msgf("Brokers=%v", conf.Brokers)
	log.Info().Msgf("ClientId=%s", conf.ClientId)
}

func (conf *MqttConfig) Validate() error {
	for _, broker := range conf.Brokers {
		if !IsValidUrl(broker) {
			return fmt.Errorf("no valid host given: %s", broker)
		}
	}

	return nil
}
