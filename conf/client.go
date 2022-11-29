//go:build client

package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"os"
)

var defaultHttpResolverUrls = []string{
	"https://icanhazip.com",
	"https://ifconfig.me",
	"https://ifconfig.co",
	"https://ipinfo.io/ip",
	"https://api.ipify.org",
	"https://ipecho.net/plain",
	"https://checkip.amazonaws.com",
}

type ClientConf struct {
	Host            string   `json:"host,omitempty" env:"DYNDNS_HOST"`
	KeyPairPath     string   `json:"keypair_path,omitempty" env:"DYNDNS_KEYPAIR_PATH"`
	MetricsListener string   `json:"metrics_listen,omitempty" env:"DYNDNS_METRICS_LISTEN"`
	Urls            []string `json:"http_resolver_urls,omitempty" env:"DYNDNS_HTTP_RESOLVER_URLS" envSeparator:";"`
	Once            bool     // this is not parsed via json, it's an cli flag
	MqttConfig
	*InterfaceConfig
}

func (conf *ClientConf) Print() {
	log.Info().Msg("Config in use:")
	log.Info().Msgf("host=%s", conf.Host)
	log.Info().Msgf("KeyPairPath=%s", conf.KeyPairPath)
	log.Info().Msgf("Once=%t", conf.Once)
	log.Info().Msgf("MetricsListener=%s", conf.MetricsListener)
	if len(conf.Urls) > 0 {
		log.Info().Msgf("Urls=%v", conf.Urls)
	}
	conf.MqttConfig.Print()
	if conf.InterfaceConfig != nil {
		conf.InterfaceConfig.Print()
	}
	log.Info().Msg("---")
}

func (conf *ClientConf) Validate() error {
	if len(conf.Host) == 0 {
		return errors.New("no host given")
	}

	if len(conf.KeyPairPath) == 0 {
		return errors.New("no path for keypair given")
	}

	if conf.InterfaceConfig != nil {
		err := conf.InterfaceConfig.Validate()
		if err != nil {
			return err
		}
	}

	return conf.MqttConfig.Validate()
}

func getDefaultClientConfig() *ClientConf {
	return &ClientConf{
		MetricsListener: metrics.DefaultListener,
		Urls:            defaultHttpResolverUrls,
	}
}

func ReadClientConfig(path string) (*ClientConf, error) {
	if path == "" {
		return &ClientConf{}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %s: %v", path, err)
	}

	conf := getDefaultClientConfig()
	err = json.Unmarshal(content, &conf)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal json to config: %v", err)
	}

	return conf, nil
}
