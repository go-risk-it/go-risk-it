package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type DatabaseConfig struct {
	Host       string `koanf:"host"`
	Port       int    `koanf:"port"`
	Name       string `koanf:"name"`
	User       string `koanf:"user"`
	Password   string `koanf:"password"`
	DisableSSL bool   `koanf:"disable_ssl"`
}

func (c *DatabaseConfig) BuildConnectionString() string {
	hostPort := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))

	result := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s",
		c.User,
		c.Password,
		hostPort,
		c.Name,
	)

	if c.DisableSSL {
		result += "?sslmode=disable"
	}

	return result
}

type JwtConfig struct {
	Secret []byte `koanf:"secret"`
}

type Config struct {
	Jwt      JwtConfig
	Database DatabaseConfig
}

type Result struct {
	fx.Out

	JwtConfig      JwtConfig
	DatabaseConfig DatabaseConfig
}

func newConfig(log *zap.SugaredLogger) Result {
	koanfManager := koanf.New(".")

	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	readFromConfigFile(koanfManager)
	readFromEnv(koanfManager)

	var config Config
	if err := koanfManager.Unmarshal("", &config); err != nil {
		panic(err)
	}

	log.Debugf("Loaded config: %+v", koanfManager)

	return Result{
		JwtConfig:      config.Jwt,
		DatabaseConfig: config.Database,
	}
}

func readFromConfigFile(k *koanf.Koanf) {
	if err := k.Load(file.Provider("local.yml"), yaml.Parser()); err != nil {
		panic(err)
	}
}

func readFromEnv(k *koanf.Koanf) {
	err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, "")), "_", ".")
	}), nil)
	if err != nil {
		panic(err)
	}
}

var Module = fx.Options(
	fx.Provide(
		newConfig,
	),
)
