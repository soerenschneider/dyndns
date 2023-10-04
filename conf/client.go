//go:build client

package conf

import (
	"encoding/json"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/metrics"
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
		"/etc/dyndns/client.json",
		"~/.dyndns/config.json",
	}
)

type ClientConf struct {
	Host             string   `json:"host,omitempty" env:"DYNDNS_HOST" validate:"required"`
	AddrFamilies     []string `json:"address_families" env:"DYNDNS_ADDRESS_FAMILIES" envSeparator:";" validate:"omitempty,addrfamilies"`
	KeyPairPath      string   `json:"keypair_path,omitempty" env:"DYNDNS_KEYPAIR_PATH" validate:"required_if=KeyPair '',omitempty,filepath"`
	KeyPair          string   `json:"keypair,omitempty" env:"DYNDNS_KEYPAIR" validate:"required_if=KeyPairPath ''"`
	MetricsListener  string   `json:"metrics_listen,omitempty" env:"DYNDNS_METRICS_LISTEN"`
	PreferredUrls    []string `json:"http_resolver_preferred_urls,omitempty" env:"DYNDNS_HTTP_RESOLVER_PREFERRED_URLS" envSeparator:";"`
	FallbackUrls     []string `json:"http_resolver_fallback_urls,omitempty" env:"DYNDNS_HTTP_RESOLVER_FALLBACK_URLS" envSeparator:";"`
	NetworkInterface string   `json:"interface,omitempty"`
	Once             bool     // this is not parsed via json, it's an cli flag

	HttpConfig struct {
		Url string `json:"url"`
	} `json:"http"`
	MqttConfig
	*EmailConfig `json:"notifications"`
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

	if err := json.Unmarshal(content, &conf); err != nil {
		return nil, err
	}

	return conf, nil
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
