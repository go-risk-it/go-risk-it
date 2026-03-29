package config

type JwtConfig struct {
	Secret []byte `koanf:"secret"`
}
