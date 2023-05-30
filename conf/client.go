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

const (
	AddrFamilyIpv6 = "ip6"
	AddrFamilyIpv4 = "ip4"
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
	Host            string   `json:"host,omitempty" env:"DYNDNS_HOST" validate:"required"`
	AddrFamilies    []string `json:"address_families" env:"ADDRESS_FAMILIES" envSeparator:";" validate:"omitempty,oneof=ipv4 ipv6"`
	KeyPairPath     string   `json:"keypair_path,omitempty" env:"DYNDNS_KEYPAIR_PATH" validate:"file"`
	MetricsListener string   `json:"metrics_listen,omitempty" env:"DYNDNS_METRICS_LISTEN"`
	PreferredUrls   []string `json:"http_resolver_preferred_urls,omitempty" env:"DYNDNS_HTTP_RESOLVER_PREFERRED_URLS" envSeparator:";"`
	FallbackUrls    []string `json:"http_resolver_fallback_urls,omitempty" env:"DYNDNS_HTTP_RESOLVER_FALLBACK_URLS" envSeparator:";"`
	Once            bool     // this is not parsed via json, it's an cli flag
	MqttConfig
	*EmailConfig `json:"notifications"`
	*InterfaceConfig
}

func (conf *ClientConf) Print() {
	log.Info().Msg("Config in use:")
	log.Info().Msgf("host=%s", conf.Host)
	log.Info().Msgf("KeyPairPath=%s", conf.KeyPairPath)
	log.Info().Msgf("Once=%t", conf.Once)
	log.Info().Msgf("MetricsListener=%s", conf.MetricsListener)
	if len(conf.PreferredUrls) > 0 {
		log.Info().Msgf("PreferredUrls=%v", conf.PreferredUrls)
	}
	if len(conf.FallbackUrls) > 0 {
		log.Info().Msgf("FallbackUrls=%v", conf.FallbackUrls)
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

	if len(conf.PreferredUrls) == 0 {
		return errors.New("no preferred urls given")
	}

	return conf.MqttConfig.Validate()
}

func getDefaultClientConfig() *ClientConf {
	return &ClientConf{
		MetricsListener: metrics.DefaultListener,
		AddrFamilies:    []string{AddrFamilyIpv4, AddrFamilyIpv6},
		PreferredUrls:   defaultHttpResolverUrls,
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
