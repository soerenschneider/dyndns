//go:build client

package conf

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"os"
	"reflect"
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
	AddrFamilies    []string `json:"address_families" env:"ADDRESS_FAMILIES" envSeparator:";" validate:"omitempty,addrfamilies"`
	KeyPairPath     string   `json:"keypair_path,omitempty" env:"DYNDNS_KEYPAIR_PATH" validate:"filepath"`
	MetricsListener string   `json:"metrics_listen,omitempty" env:"DYNDNS_METRICS_LISTEN"`
	PreferredUrls   []string `json:"http_resolver_preferred_urls,omitempty" env:"DYNDNS_HTTP_RESOLVER_PREFERRED_URLS" envSeparator:";"`
	FallbackUrls    []string `json:"http_resolver_fallback_urls,omitempty" env:"DYNDNS_HTTP_RESOLVER_FALLBACK_URLS" envSeparator:";"`
	Once            bool     // this is not parsed via json, it's an cli flag
	MqttConfig
	*EmailConfig `json:"notifications"`
	*InterfaceConfig
}

func (c *ClientConf) Print() {
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
