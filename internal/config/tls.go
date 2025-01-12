package config

type TLSConfig struct {
	Cert               []byte `koanf:"cert"`
	Key                []byte `koanf:"key"`
	InsecureSkipVerify bool   `koanf:"insecure_skip_verify"`
}
