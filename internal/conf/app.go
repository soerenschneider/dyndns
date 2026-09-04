package conf

import (
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
)

type Conf struct {
	Client *ClientConf `yaml:"client" validate:"required_without=Mode server"`
	Server *ServerConf `yaml:"server" validate:"required_without=Mode client"`
	Mode   string      `yaml:"mode" validate:"oneof=client server edge"`

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
