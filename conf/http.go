package conf

type HttpConfig struct {
	ListenAddr string `yaml:"addr"`
	TlsCert    string `yaml:"tls_cert" validate:"required_with_all=TlsKey,omitempty,file"`
	TlsKey     string `yaml:"tls_key" validate:"required_with_all=TlsCert,omitempty,file"`
}
