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

type Config struct {
	Jwt      JwtConfig
	Database DatabaseConfig
	Dice     DiceConfig
	History  HistoryConfig
}

type Result struct {
	fx.Out

	JwtConfig      JwtConfig
	DatabaseConfig DatabaseConfig
	DiceConfig     DiceConfig
	HistoryConfig  HistoryConfig
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
		DiceConfig:     config.Dice,
		HistoryConfig:  config.History,
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
