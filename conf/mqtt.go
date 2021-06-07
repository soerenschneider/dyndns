package conf

import (
	"fmt"
	"github.com/rs/zerolog/log"
)

type MqttConfig struct {
	Broker   string `json:"broker"`
	ClientId string `json:"client_id"`
}

func (conf *MqttConfig) Print() {
	log.Info().Msgf("Broker=%s", conf.Broker)
	log.Info().Msgf("ClientId=%s", conf.ClientId)
}

func (conf *MqttConfig) Validate() error {
	if !IsValidUrl(conf.Broker) {
		return fmt.Errorf("no valid host given: %s", conf.Broker)
	}

	return nil
}
