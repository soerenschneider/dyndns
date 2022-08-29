//go:build server

package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/verification"
	"io/ioutil"
)

type ServerConf struct {
	KnownHosts      map[string][]string `json:"known_hosts"`
	HostedZoneId    string              `json:"hosted_zone_id"`
	MetricsListener string              `json:"metrics_listen",omitempty`
	MqttConfig
	VaultConfig
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

func (conf *ServerConf) Validate() error {
	if len(conf.KnownHosts) == 0 {
		return errors.New("no hosts configured")
	}

	if len(conf.HostedZoneId) == 0 {
		return errors.New("no hosted zone id provided")
	}

	return conf.MqttConfig.Validate()
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
					ret[host] = make([]verification.VerificationKey, len(configuredPubkeys))
				}
				ret[host] = append(ret[host], publicKey)
			}
		}
	}

	return ret
}

func (conf *ServerConf) Print() {
	log.Info().Msgf("Configured %d hosts", len(conf.KnownHosts))
	for host, pubKey := range conf.KnownHosts {
		log.Info().Msgf("%s with pubKey %s", host, pubKey)
	}
	log.Info().Msgf("HostedZoneId=%s", conf.HostedZoneId)
	log.Info().Msgf("MetricsListener=%s", conf.MetricsListener)
	conf.MqttConfig.Print()
	conf.VaultConfig.Print()
}
