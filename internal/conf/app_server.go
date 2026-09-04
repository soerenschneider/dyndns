package conf

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"reflect"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/internal/verification"
	"gopkg.in/yaml.v3"
)

type ServerConf struct {
	KnownHosts         map[string][]string `yaml:"known_hosts" env:"KNOWN_HOSTS"`
	HostedZoneId       string              `yaml:"hosted_zone_id" env:"HOSTED_ZONE_ID"`
	hybrid.SqsConfig   `yaml:"sqs"`
	HttpConfig         `yaml:"http"`
	hybrid.MqttConfig  `yaml:"mqtt"`
	hybrid.EmailConfig `yaml:"notifications"`
	hybrid.NatsConfig  `yaml:"nats" envPrefix:"NATS_"`
}

func GetDefaultServerConfig() *ServerConf {
	return &ServerConf{
		SqsConfig: hybrid.DefaultSqsConfig(),
		MqttConfig: hybrid.MqttConfig{
			ClientId: "dyndns-server",
		},
	}
}

func ReadConfig(path string) (*Conf, error) {
	conf := GetDefaultConfig()
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

func ParseEnvVariables(config *Conf) error {
	funk := map[reflect.Type]env.ParserFunc{}

	funk[reflect.TypeOf(map[string][]string{})] = func(input string) (any, error) {
		var ret = map[string][]string{}
		return ret, json.Unmarshal([]byte(input), &ret)
	}

	funk[reflect.TypeOf([]HttpDispatcherConfig{})] = func(input string) (any, error) {
		var ret []HttpDispatcherConfig
		return ret, json.Unmarshal([]byte(input), &ret)
	}

	opts := env.Options{
		Prefix: "DYNDNS_",
	}

	return env.ParseWithFuncs(config, funk, opts)
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
