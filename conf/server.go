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
	"gopkg.in/yaml.v3"
)

type ServerConf struct {
	KnownHosts      map[string][]string `yaml:"known_hosts" env:"KNOWN_HOSTS" validate:"required"`
	HostedZoneId    string              `yaml:"hosted_zone_id" env:"HOSTED_ZONE_ID" validate:"required"`
	MetricsListener string              `yaml:"metrics_listen,omitempty" validate:"omitempty,tcp_addr"`
	SqsQueue        string              `yaml:"sqs_queue" env:"SQS_QUEUE"`
	HttpConfig      `yaml:"http"`
	MqttConfig      `yaml:"mqtt"`
	VaultConfig     `yaml:"vault"`
	EmailConfig     `yaml:"notifications"`
}

func GetDefaultServerConfig() *ServerConf {
	return &ServerConf{
		MetricsListener: metrics.DefaultListener,
		MqttConfig: MqttConfig{
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

	if err := yaml.Unmarshal(content, &conf); err != nil {
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

	opts := env.Options{
		Prefix: "DYNDNS_",
	}

	return env.ParseWithFuncs(serverConf, funk, opts)
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
