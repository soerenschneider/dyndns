package conf

import (
	"github.com/soerenschneider/dyndns/v2/internal"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
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
)

type ClientConf struct {
	Host         string   `yaml:"host,omitempty" env:"HOST" validate:"required_for_mode=client edge"`
	AddrFamilies []string `yaml:"address_families" env:"ADDRESS_FAMILIES" envSeparator:";" validate:"omitempty,addrfamilies"`
	// KeyPairPath and KeyPair are alternatives, KeyPair takes precedence.
	KeyPairPath      string   `yaml:"keypair_path,omitempty" env:"KEYPAIR_PATH"`
	KeyPair          string   `yaml:"keypair,omitempty" env:"KEYPAIR"`
	PreferredUrls    []string `yaml:"http_resolver_preferred_urls,omitempty" env:"HTTP_RESOLVER_PREFERRED_URLS" envSeparator:";"`
	FallbackUrls     []string `yaml:"http_resolver_fallback_urls,omitempty" env:"HTTP_RESOLVER_FALLBACK_URLS" envSeparator:";"`
	NetworkInterface string   `yaml:"interface,omitempty"`
	// Ipv4Resolver and Ipv6Resolver pick the mechanism per address family, e.g.
	// the public IPv4 via http behind NAT and the IPv6 straight from the interface.
	Ipv4Resolver string `yaml:"ipv4_resolver,omitempty" env:"IPV4_RESOLVER" validate:"omitempty,oneof=http interface"`
	Ipv6Resolver string `yaml:"ipv6_resolver,omitempty" env:"IPV6_RESOLVER" validate:"omitempty,oneof=http interface"`
	// Once and ForceSendUpdate are set via cli flags, not via the config file.
	Once            bool `yaml:"-"`
	ForceSendUpdate bool `yaml:"-"`

	HttpDispatcherConf []HttpDispatcherConfig `yaml:"http_dispatcher" env:"HTTP_DISPATCHER_CONF"`
	hybrid.SqsConfig   `yaml:"sqs" envPrefix:"SQS_"`
	Aws                hybrid.AwsCredentialsConfig `yaml:"aws" envPrefix:"AWS_"`
	hybrid.MqttConfig  `yaml:"mqtt"`
	hybrid.NatsConfig  `yaml:"nats" envPrefix:"NATS_"`
}

type HttpDispatcherConfig struct {
	Url string `yaml:"url"`
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
