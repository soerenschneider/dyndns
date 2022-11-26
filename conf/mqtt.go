package conf

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

type MqttConfig struct {
	Brokers        []string `json:"brokers" env:"DYNDNS_BROKERS" envSeparator:";"`
	ClientId       string   `json:"client_id" env:"DYNDNS_CLIENT_ID"`
	CaCertFile     string   `json:"tls_ca_cert" env:"DYNDNS_TLS_CA"`
	ClientCertFile string   `json:"tls_client_cert" env:"DYNDNS_TLS_CERT"`
	ClientKeyFile  string   `json:"tls_client_key" env:"DYNDNS_TLS_KEY"`
	TlsInsecure    bool     `json:"tls_insecure" env:"DYNDNS_TLS_INSECURE"`
}

func (conf *MqttConfig) TlsConfig() *tls.Config {
	log.Info().Msg("Building TLS config...")

	certPool, err := x509.SystemCertPool()
	if err != nil {
		log.Warn().Msgf("Could not get system cert pool")
		if len(conf.CaCertFile) > 0 {
			log.Warn().Msgf("Creating empty cert pool")
			certPool = x509.NewCertPool()
		}
	}

	pemCerts, err := os.ReadFile(conf.CaCertFile)
	if err != nil {
		log.Error().Msgf("Could not read CA cert file: %v", err)
	} else {
		certPool.AppendCertsFromPEM(pemCerts)
	}

	var certs []tls.Certificate
	clientCertDefined := len(conf.ClientCertFile) > 0
	clientKeyDefined := len(conf.ClientKeyFile) > 0
	if clientCertDefined && clientKeyDefined {
		cert, err := tls.LoadX509KeyPair(conf.ClientCertFile, conf.ClientKeyFile)
		if err != nil {
			log.Panic().Msgf("could not read tls key pair: %v", err)
		}

		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			log.Panic().Msgf("could not parse tls key pair: %v", err)
		}

		certs = []tls.Certificate{cert}
	}

	return &tls.Config{
		RootCAs:            certPool,
		ClientAuth:         tls.RequestClientCert,
		Certificates:       certs,
		ClientCAs:          nil,
		InsecureSkipVerify: conf.TlsInsecure,
	}
}

func (conf *MqttConfig) Print() {
	log.Info().Msgf("Brokers=%v", conf.Brokers)
	log.Info().Msgf("ClientId=%s", conf.ClientId)
	if len(conf.CaCertFile) > 1 {
		log.Info().Msgf("CaCertFile=%s", conf.CaCertFile)
	}
	if len(conf.ClientCertFile) > 1 {
		log.Info().Msgf("ClientCertFile=%s", conf.ClientCertFile)
	}
	if len(conf.ClientKeyFile) > 1 {
		log.Info().Msgf("ClientKeyFile=%s", conf.ClientKeyFile)
	}
}

func (conf *MqttConfig) Validate() error {
	for _, broker := range conf.Brokers {
		if !IsValidUrl(broker) {
			return fmt.Errorf("no valid host given: %s", broker)
		}
	}

	return nil
}
