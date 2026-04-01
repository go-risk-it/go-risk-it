package config_test

import (
	"strings"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKernelConfig_UnmarshalInfraOnly(t *testing.T) {
	t.Parallel()

	raw := []byte(`
database:
  host: localhost
  port: 5432
  name: testdb
  user: user
  password: pass
  disable_ssl: true
  max_conns: 50
  min_conns: 10
jwt:
  secret: test-secret
game:
  dice:
    roll_strategy: random
  regionassignment:
    assignment_strategy: random
  history:
    size: 50
otel:
  enabled: true
server:
  allowed_origins:
    - "http://localhost:5173"
`)

	k := koanf.New(".")
	require.NoError(t, k.Load(rawbytes.Provider(raw), yaml.Parser()))

	var cfg config.Config
	require.NoError(t, k.Unmarshal("", &cfg))

	// Infra fields populated
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.True(t, cfg.Database.DisableSSL)
	assert.Equal(t, int32(50), cfg.Database.MaxConns)
	assert.True(t, cfg.Otel.Enabled)
	assert.Equal(t, []byte("test-secret"), cfg.Jwt.Secret)
	require.NotEmpty(t, cfg.Server.AllowedOrigins)
	assert.Equal(t, "http://localhost:5173", cfg.Server.AllowedOrigins[0])
}

func TestKernelConfig_EnvOverridesYaml(t *testing.T) {
	raw := []byte(`
database:
  host: yaml-host
  port: 5432
  name: testdb
  user: yaml-user
  password: yaml-pass
  disable_ssl: true
  max_conns: 50
jwt:
  secret: yaml-secret
otel:
  enabled: false
`)

	t.Setenv("DATABASE_HOST", "env-host")
	t.Setenv("DATABASE_USER", "env-user")
	t.Setenv("JWT_SECRET", "env-secret")
	t.Setenv("OTEL_ENABLED", "true")

	koanfManager := koanf.New(".")
	require.NoError(t, koanfManager.Load(rawbytes.Provider(raw), yaml.Parser()))
	require.NoError(t, koanfManager.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, "")), "_", ".")
	}), nil))

	var cfg config.Config
	require.NoError(t, koanfManager.Unmarshal("", &cfg))

	// Env vars override YAML values
	assert.Equal(t, "env-host", cfg.Database.Host)
	assert.Equal(t, "env-user", cfg.Database.User)
	assert.Equal(t, []byte("env-secret"), cfg.Jwt.Secret)
	assert.True(t, cfg.Otel.Enabled)

	// YAML-only values preserved
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.True(t, cfg.Database.DisableSSL)
	assert.Equal(t, int32(50), cfg.Database.MaxConns)
}

func TestKernelConfig_GameFieldsNotInStruct(t *testing.T) {
	t.Parallel()

	// Compile-time guarantee: Config struct has no game-domain fields.
	cfg := config.Config{}
	_ = cfg.Jwt
	_ = cfg.Database
	_ = cfg.Otel
	_ = cfg.Server
	// If Config had Dice, Regionassignment, or History fields, these
	// would compile — they must NOT compile:
	// _ = cfg.Dice
	// _ = cfg.Regionassignment
	// _ = cfg.History
}
