package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/yaml"
	env "github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
)

const EnvironmentKey = "ENVIRONMENT"

// Config holds infrastructure-only configuration (no game-domain sections).
type Config struct {
	Jwt      JwtConfig
	Database DatabaseConfig
	Otel     OtelConfig
	Server   ServerConfig
}

// Result provides individual infrastructure config types as fx dependencies.
type Result struct {
	fx.Out

	JwtConfig      JwtConfig
	DatabaseConfig DatabaseConfig
	OtelConfig     OtelConfig
	ServerConfig   ServerConfig
}

// newKoanf creates and loads the koanf configuration manager. The returned
// *koanf.Koanf is provided to downstream modules (e.g. game/logic/config)
// so they can unmarshal their own sections.
func newKoanf() (*koanf.Koanf, error) {
	koanfManager := koanf.New(".")

	err := godotenv.Load(".env")
	if err != nil {
		slog.Debug("failed to load .env")
	}

	if err := readFromConfigFile(koanfManager); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := readFromEnv(koanfManager); err != nil {
		return nil, fmt.Errorf("failed to read env config: %w", err)
	}

	slog.Debug("loaded config", "config", fmt.Sprintf("%+v", koanfManager))

	return koanfManager, nil
}

// newConfig unmarshals the top-level infrastructure config from koanf.
func newConfig(k *koanf.Koanf) (Result, error) {
	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		return Result{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return Result{
		JwtConfig:      config.Jwt,
		DatabaseConfig: config.Database,
		OtelConfig:     config.Otel,
		ServerConfig:   config.Server,
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

// TransformKey converts environment variable names to koanf-compatible
// dotted key paths. For example, DATABASE_HOST becomes database.host.
// Multi-word YAML keys (e.g. disable_ssl) cannot be set via env vars
// because every underscore becomes a dot separator.
//
// Exported for use in test code that constructs its own env.Provider
// with a custom EnvironFunc.
func TransformKey(key, value string) (string, any) {
	return strings.ReplaceAll(strings.ToLower(key), "_", "."), value
}

func readFromEnv(k *koanf.Koanf) error {
	err := k.Load(env.Provider(".", env.Opt{
		TransformFunc: TransformKey,
	}), nil)
	if err != nil {
		return fmt.Errorf("failed to load env vars: %w", err)
	}

	return nil
}

var Module = fx.Options(
	fx.Provide(
		newKoanf,
		newConfig,
	),
)
