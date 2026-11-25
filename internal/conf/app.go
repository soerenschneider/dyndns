package conf

import (
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
)

type Conf struct {
	// Both are needed in every mode, e.g. the server mode reads the cli flags
	// from Client, so an empty "client:" or "server:" section is rejected.
	Client *ClientConf `yaml:"client" validate:"required"`
	Server *ServerConf `yaml:"server" validate:"required"`
	// Mode is chosen by the subcommand, not by the config file.
	Mode string `yaml:"-" validate:"oneof=client server edge"`

	MetricsListener    string `yaml:"metrics_listen,omitempty" validate:"omitempty,tcp_addr"`
	hybrid.EmailConfig `yaml:"notifications"`
}

func GetDefaultConfig() *Conf {
	return &Conf{
		Server:          GetDefaultServerConfig(),
		Client:          getDefaultClientConfig(),
		MetricsListener: metrics.DefaultListener,
	}
}
