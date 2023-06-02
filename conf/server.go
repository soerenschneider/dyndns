//go:build server

package conf

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/verification"
	"io/ioutil"
	"reflect"
)

type ServerConf struct {
	KnownHosts      map[string][]string `json:"known_hosts" validate:"required"`
	HostedZoneId    string              `json:"hosted_zone_id" validate:"required"`
	MetricsListener string              `json:"metrics_listen",omitempty`
	MqttConfig
	VaultConfig
	*EmailConfig `json:"notifications"`
}

func (c *ServerConf) Print() {
	log.Info().Msg("---")
	log.Info().Msg("Active config values:")
	val := reflect.ValueOf(c).Elem()
	for i := 0; i < val.NumField(); i++ {
		if !val.Field(i).IsZero() {
			fieldName := val.Type().Field(i).Tag.Get("mapstructure")
			log.Info().Msgf("%s=%v", fieldName, val.Field(i))
		}
	}
	log.Info().Msg("---")
}

func getDefaultServerConfig() *ServerConf {
	return &ServerConf{
		MetricsListener: metrics.DefaultListener,
		MqttConfig: MqttConfig{
			ClientId: "dyndns-server",
		},
		VaultConfig: GetDefaultVaultConfig(),
	}
}

func ReadServerConfig(path string) (*ServerConf, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %s: %v", path, err)
	}

	conf := getDefaultServerConfig()
	err = json.Unmarshal(content, &conf)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal json to config: %v", err)
	}

	return conf, nil
}

func (conf *ServerConf) DecodePublicKeys() map[string][]verification.VerificationKey {
	var ret = map[string][]verification.VerificationKey{}

	for host, configuredPubkeys := range conf.KnownHosts {
		if len(configuredPubkeys) == 0 {
			metrics.PublicKeyErrors.Inc()
			log.Info().Msgf("No publickey defined for host %s", host)
			continue
		}

		for i, key := range configuredPubkeys {
			publicKey, err := verification.PubkeyFromString(key)
			if err != nil {
				metrics.PublicKeyErrors.Inc()
				log.Info().Msgf("Could not initialize %d. publicKey for host %s: %v", i, host, err)
			} else {
				if ret[host] == nil {
					ret[host] = make([]verification.VerificationKey, 0, len(configuredPubkeys))
				}
				ret[host] = append(ret[host], publicKey)
			}
		}
	}

	return ret
}
