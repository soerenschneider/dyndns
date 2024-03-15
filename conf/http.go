package conf

type HttpConfig struct {
	ListenAddr string `yaml:"addr" validate:""`
	TlsCert    string `yaml:"tls_cert" validate:"required_with=TlsKey,file"`
	TlsKey     string `yaml:"tls_key" validate:"required_with=TlsCert,file"`
}
