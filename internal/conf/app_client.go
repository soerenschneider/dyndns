package conf

import (
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"gopkg.in/yaml.v3"
)

var (
	defaultHttpResolverUrls = []string{
		"https://icanhazip.com",
		"https://ifconfig.me",
		"https://ident.me",
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
	KeyPairPath      string   `yaml:"keypair_path,omitempty" env:"KEYPAIR_PATH" validate:"required_for_mode=client"`
	KeyPair          string   `yaml:"keypair,omitempty" env:"KEYPAIR" validate:"required_for_mode=client"`
	MetricsListener  string   `yaml:"metrics_listen,omitempty" env:"METRICS_LISTEN"`
	PreferredUrls    []string `yaml:"http_resolver_preferred_urls,omitempty" env:"HTTP_RESOLVER_PREFERRED_URLS" envSeparator:";"`
	FallbackUrls     []string `yaml:"http_resolver_fallback_urls,omitempty" env:"HTTP_RESOLVER_FALLBACK_URLS" envSeparator:";"`
	NetworkInterface string   `yaml:"interface,omitempty"`
	// Ipv4Resolver and Ipv6Resolver pick the mechanism per address family, e.g.
	// the public IPv4 via http behind NAT and the IPv6 straight from the interface.
	Ipv4Resolver string `yaml:"ipv4_resolver,omitempty" env:"IPV4_RESOLVER" validate:"omitempty,oneof=http interface"`
	Ipv6Resolver string `yaml:"ipv6_resolver,omitempty" env:"IPV6_RESOLVER" validate:"omitempty,oneof=http interface"`
	Once         bool   // this is not parsed via json, it's an cli flag

	HttpDispatcherConf []HttpDispatcherConfig `yaml:"http_dispatcher" env:"HTTP_DISPATCHER_CONF"`
	hybrid.SqsConfig   `yaml:"sqs" envPrefix:"SQS_"`
	hybrid.MqttConfig  `yaml:"mqtt"`
	hybrid.NatsConfig  `yaml:"nats" envPrefix:"NATS_"`
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

func getDefaultClientConfig() *ClientConf {
	return &ClientConf{
		SqsConfig:     hybrid.DefaultSqsConfig(),
		AddrFamilies:  []string{internal.AddrFamilyIpv4},
		PreferredUrls: defaultHttpResolverUrls,
	}
}

// ResolverFor returns the resolver mechanism used for addrFamily. Unless set
// explicitly, the interface resolver is used if an interface is configured and
// the http resolver otherwise.
func (c *ClientConf) ResolverFor(addrFamily string) string {
	var explicit string
	switch addrFamily {
	case internal.AddrFamilyIpv4:
		explicit = c.Ipv4Resolver
	case internal.AddrFamilyIpv6:
		explicit = c.Ipv6Resolver
	}

	if explicit != "" {
		return explicit
	}

	if c.NetworkInterface != "" {
		return internal.ResolverInterface
	}

	return internal.ResolverHttp
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
