package conf

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/rs/zerolog/log"
)

type MqttConfig struct {
	Brokers        []string `yaml:"brokers" env:"BROKERS" envSeparator:";" validate:"broker"`
	ClientId       string   `yaml:"client_id" env:"CLIENT_ID" validate:"required_with=Brokers ''"`
	CaCertFile     string   `yaml:"tls_ca_cert" env:"TLS_CA" validate:"omitempty,file"`
	ClientCertFile string   `yaml:"tls_client_cert" env:"TLS_CERT" validate:"omitempty,required_unless=ClientKeyFile '',file"`
	ClientKeyFile  string   `yaml:"tls_client_key" env:"TLS_KEY" validate:"omitempty,required_unless=ClientCertFile '',file"`
	TlsInsecure    bool     `yaml:"tls_insecure" env:"TLS_INSECURE"`
}

func (conf *MqttConfig) UsesTlsClientCerts() bool {
	return len(conf.CaCertFile) > 0 && len(conf.ClientCertFile) > 0 && len(conf.ClientKeyFile) > 0
}

func (conf *MqttConfig) TlsConfig() *tls.Config {
	log.Info().Str("component", "config").Msg("Building TLS config...")

	certPool, err := x509.SystemCertPool()
	if err != nil {
		log.Warn().Err(err).Str("component", "config").Msg("Could not get system cert pool")
		certPool = x509.NewCertPool()
	}

	if conf.UsesTlsClientCerts() {
		pemCerts, err := os.ReadFile(conf.CaCertFile)
		if err != nil {
			log.Error().Err(err).Str("component", "config").Msg("Could not read CA cert file")
		} else {
			certPool.AppendCertsFromPEM(pemCerts)
		}
	}

	tlsConf := &tls.Config{
		RootCAs:            certPool,
		ClientAuth:         tls.RequestClientCert,
		InsecureSkipVerify: conf.TlsInsecure, // #nosec G402
	}

	clientCertDefined := len(conf.ClientCertFile) > 0
	clientKeyDefined := len(conf.ClientKeyFile) > 0
	if clientCertDefined && clientKeyDefined {
		tlsConf.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := tls.LoadX509KeyPair(conf.ClientCertFile, conf.ClientKeyFile)
			return &cert, err
		}
	}

	return tlsConf
}
