package config

type ServerConfig struct {
	AllowedOrigins []string `koanf:"allowed_origins"`
}
