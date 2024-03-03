//go:build client

package conf

import (
	"encoding/json"
	"os"
	"os/user"
	"path"
	"reflect"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"gopkg.in/yaml.v3"
)

var (
	defaultHttpResolverUrls = []string{
		"https://icanhazip.com",
		"https://ifconfig.me",
		"https://ifconfig.co",
		"https://ipinfo.io/ip",
		"https://api.ipify.org",
		"https://ipecho.net/plain",
		"https://checkip.amazonaws.com",
	}

	configPathPreferences = []string{
		"/etc/dyndns/client.yaml",
		"~/.dyndns/config.yaml",
	}
)

type ClientConf struct {
	Host             string   `yaml:"host,omitempty" env:"HOST" validate:"required"`
	AddrFamilies     []string `yaml:"address_families" env:"ADDRESS_FAMILIES" envSeparator:";" validate:"omitempty,addrfamilies"`
	KeyPairPath      string   `yaml:"keypair_path,omitempty" env:"KEYPAIR_PATH" validate:"required_if=KeyPair '',omitempty,filepath"`
	KeyPair          string   `yaml:"keypair,omitempty" env:"KEYPAIR" validate:"required_if=KeyPairPath ''"`
	MetricsListener  string   `yaml:"metrics_listen,omitempty" env:"METRICS_LISTEN"`
	PreferredUrls    []string `yaml:"http_resolver_preferred_urls,omitempty" env:"HTTP_RESOLVER_PREFERRED_URLS" envSeparator:";"`
	FallbackUrls     []string `yaml:"http_resolver_fallback_urls,omitempty" env:"HTTP_RESOLVER_FALLBACK_URLS" envSeparator:";"`
	NetworkInterface string   `yaml:"interface,omitempty"`
	Once             bool     // this is not parsed via json, it's an cli flag

	HttpDispatcherConf []HttpDispatcherConfig `yaml:"http_dispatcher" env:"HTTP_DISPATCHER_CONF"`
	*MqttConfig        `yaml:"mqtt"`
	*EmailConfig       `yaml:"notifications"`
}

type HttpDispatcherConfig struct {
	Url string `yaml:"url"`
}

func ReadClientConfig(path string) (*ClientConf, error) {
	conf := getDefaultClientConfig()
	if path == "" {
		return conf, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(content, &conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func ParseClientConfEnv(clientConf *ClientConf) error {
	funk := map[reflect.Type]env.ParserFunc{}

	funk[reflect.TypeOf([]HttpDispatcherConfig{})] = func(input string) (any, error) {
		var ret []HttpDispatcherConfig
		return ret, json.Unmarshal([]byte(input), &ret)
	}

	opts := env.Options{
		Prefix: "DYNDNS_",
	}

	return env.ParseWithFuncs(clientConf, funk, opts)
}

func getDefaultClientConfig() *ClientConf {
	return &ClientConf{
		MetricsListener: metrics.DefaultListener,
		AddrFamilies:    []string{AddrFamilyIpv4},
		PreferredUrls:   defaultHttpResolverUrls,
	}
}

func GetDefaultConfigFileOrEmpty() string {
	homeDir := getUserHomeDirectory()
	for _, configPath := range configPathPreferences {
		if homeDir != "" {
			if strings.HasPrefix(configPath, "~/") {
				configPath = path.Join(homeDir, configPath[2:])
			} else if strings.HasPrefix(configPath, "$HOME/") {
				configPath = path.Join(homeDir, configPath[6:])
			}
		}

		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return ""
}

func getUserHomeDirectory() string {
	usr, err := user.Current()
	if err != nil || usr == nil {
		log.Warn().Msg("Could not find user home directory")
		return ""
	}
	dir := usr.HomeDir
	return dir
}
