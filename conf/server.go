//go:build server

package conf

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"reflect"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"github.com/soerenschneider/dyndns/internal/verification"
)

type ServerConf struct {
	KnownHosts      map[string][]string `json:"known_hosts" env:"DYNDNS_KNOWN_HOSTS" validate:"required"`
	HostedZoneId    string              `json:"hosted_zone_id" env:"DYNDNS_HOSTED_ZONE_ID" validate:"required"`
	MetricsListener string              `json:"metrics_listen,omitempty"`
	*MqttConfig
	*VaultConfig
	*EmailConfig `json:"notifications"`
}

func GetDefaultServerConfig() *ServerConf {
	return &ServerConf{
		MetricsListener: metrics.DefaultListener,
		MqttConfig: &MqttConfig{
			ClientId: "dyndns-server",
		},
		VaultConfig: GetDefaultVaultConfig(),
	}
}

func ReadServerConfig(path string) (*ServerConf, error) {
	conf := GetDefaultServerConfig()
	if len(path) == 0 {
		return conf, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %s: %v", path, err)
	}

	if err := json.Unmarshal(content, &conf); err != nil {
		return nil, fmt.Errorf("could not unmarshal json to config: %v", err)
	}

	return conf, nil
}

func ParseEnvVariables(serverConf *ServerConf) error {
	funk := map[reflect.Type]env.ParserFunc{}

	funk[reflect.TypeOf(map[string][]string{})] = func(input string) (any, error) {
		var ret = map[string][]string{}
		return ret, json.Unmarshal([]byte(input), &ret)
	}

	return env.ParseWithFuncs(serverConf, funk)
}

func (conf *ServerConf) DecodePublicKeys() (map[string][]verification.VerificationKey, error) {
	var ret = map[string][]verification.VerificationKey{}

	for host, configuredPubkeys := range conf.KnownHosts {
		if len(configuredPubkeys) == 0 {
			log.Info().Msgf("No publickey defined for host %s", host)
			continue
		}

		for _, key := range configuredPubkeys {
			publicKey, err := verification.PubkeyFromString(key)
			if err != nil {
				return nil, fmt.Errorf("could not read pubkey: %w", err)
			}

			if ret[host] == nil {
				ret[host] = make([]verification.VerificationKey, 0, len(configuredPubkeys))
			}
			ret[host] = append(ret[host], publicKey)
		}
	}

	return ret, nil
}

func GetKnownHostsHash(knownHosts map[string][]string) (uint64, error) {
	jsonBytes, err := json.Marshal(knownHosts)
	if err != nil {
		return 0, err
	}

	hash := fnv.New64a()
	hash.Write(jsonBytes)
	return hash.Sum64(), nil
}
