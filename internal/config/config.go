package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const EnvironmentKey = "ENVIRONMENT"

type Config struct {
	Jwt              JwtConfig
	Database         DatabaseConfig
	Dice             DiceConfig
	Regionassignment RegionassignmentConfig
	History          HistoryConfig
	Otel             OtelConfig
	Server           ServerConfig
}

type Result struct {
	fx.Out

	JwtConfig              JwtConfig
	DatabaseConfig         DatabaseConfig
	DiceConfig             DiceConfig
	RegionassignmentConfig RegionassignmentConfig
	HistoryConfig          HistoryConfig
	OtelConfig             OtelConfig
	ServerConfig           ServerConfig
}

func newConfig(log *zap.SugaredLogger) (Result, error) {
	koanfManager := koanf.New(".")

	err := godotenv.Load(".env")
	if err != nil {
		log.Debugw("failed to load .env")
	}

	if err := readFromConfigFile(koanfManager); err != nil {
		return Result{}, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := readFromEnv(koanfManager); err != nil {
		return Result{}, fmt.Errorf("failed to read env config: %w", err)
	}

	var config Config
	if err := koanfManager.Unmarshal("", &config); err != nil {
		return Result{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	log.Debugf("Loaded config: %+v", koanfManager)

	return Result{
		JwtConfig:              config.Jwt,
		DatabaseConfig:         config.Database,
		DiceConfig:             config.Dice,
		RegionassignmentConfig: config.Regionassignment,
		HistoryConfig:          config.History,
		OtelConfig:             config.Otel,
		ServerConfig:           config.Server,
	}, nil
}

func readFromConfigFile(k *koanf.Koanf) error {
	if err := k.Load(file.Provider(getEnv()+".yml"), yaml.Parser()); err != nil {
		return fmt.Errorf("failed to load config file %s.yml: %w", getEnv(), err)
	}

	return nil
}

func getEnv() string {
	environment := os.Getenv(EnvironmentKey)
	if environment == "" {
		environment = "component-test"
	}

	return environment
}

func readFromEnv(k *koanf.Koanf) error {
	err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, "")), "_", ".")
	}), nil)
	if err != nil {
		return fmt.Errorf("failed to load env vars: %w", err)
	}

	return nil
}

var Module = fx.Options(
	fx.Provide(
		newConfig,
	),
)
