package config

import (
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

type Config struct {
	Database DatabaseConfig
}

type Result struct {
	fx.Out

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

	log.Infof("Loaded config: %+v", koanfManager)

	var config Config
	if err := koanfManager.Unmarshal("", &config); err != nil {
		panic(err)
	}

	log.Infof("Loaded actual config: %+v", config)

	return Result{
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
