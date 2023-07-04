package conf

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

type MqttConfig struct {
	Brokers        []string `json:"brokers" env:"DYNDNS_BROKERS" envSeparator:";" validate:"required,broker"`
	ClientId       string   `json:"client_id" env:"DYNDNS_CLIENT_ID" validate:"required"`
	CaCertFile     string   `json:"tls_ca_cert" env:"DYNDNS_TLS_CA" validate:"omitempty,file"`
	ClientCertFile string   `json:"tls_client_cert" env:"DYNDNS_TLS_CERT" validate:"omitempty,file"`
	ClientKeyFile  string   `json:"tls_client_key" env:"DYNDNS_TLS_KEY" validate:"omitempty,file"`
	TlsInsecure    bool     `json:"tls_insecure" env:"DYNDNS_TLS_INSECURE"`
}

func (conf *MqttConfig) UsesTlsClientCerts() bool {
	return len(conf.CaCertFile) > 0 && len(conf.ClientCertFile) > 0 && len(conf.ClientKeyFile) > 0
}

func (conf *MqttConfig) TlsConfig() *tls.Config {
	log.Info().Msg("Building TLS config...")

	certPool, err := x509.SystemCertPool()
	if err != nil {
		log.Warn().Msgf("Could not get system cert pool")
		certPool = x509.NewCertPool()
	}

	if conf.UsesTlsClientCerts() {
		pemCerts, err := os.ReadFile(conf.CaCertFile)
		if err != nil {
			log.Error().Msgf("Could not read CA cert file: %v", err)
		} else {
			certPool.AppendCertsFromPEM(pemCerts)
		}
	}

	tlsConf := &tls.Config{
		RootCAs:            certPool,
		ClientAuth:         tls.RequestClientCert,
		InsecureSkipVerify: conf.TlsInsecure,
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

func (conf MqttConfig) String() string {
	base := fmt.Sprintf("brokers=%v, clientId=%s", conf.Brokers, conf.ClientId)
	if conf.UsesTlsClientCerts() {
		base += fmt.Sprintf("ca=%s, crt=%s, key=%s", conf.CaCertFile, conf.ClientCertFile, conf.ClientKeyFile)
	}

	return base
}
