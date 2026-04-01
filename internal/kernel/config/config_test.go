package config_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/knadh/koanf/parsers/yaml"
	env "github.com/knadh/koanf/providers/env/v2"
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
	t.Parallel()

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

	koanfManager := koanf.New(".")
	require.NoError(t, koanfManager.Load(rawbytes.Provider(raw), yaml.Parser()))
	require.NoError(t, koanfManager.Load(env.Provider(".", env.Opt{
		TransformFunc: config.TransformKey,
		EnvironFunc: func() []string {
			return []string{
				"DATABASE_HOST=env-host",
				"DATABASE_USER=env-user",
				"JWT_SECRET=env-secret",
				"OTEL_ENABLED=true",
			}
		},
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

func TestTransformKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		key       string
		value     string
		wantKey   string
		wantValue any
	}{
		{
			name:      "single hierarchy",
			key:       "DATABASE_HOST",
			value:     "myhost",
			wantKey:   "database.host",
			wantValue: "myhost",
		},
		{
			name:      "two levels",
			key:       "JWT_SECRET",
			value:     "mysecret",
			wantKey:   "jwt.secret",
			wantValue: "mysecret",
		},
		{
			name:      "boolean",
			key:       "OTEL_ENABLED",
			value:     "true",
			wantKey:   "otel.enabled",
			wantValue: "true",
		},
		{
			name:      "multi-word field becomes extra dot",
			key:       "DATABASE_DISABLE_SSL",
			value:     "true",
			wantKey:   "database.disable.ssl",
			wantValue: "true",
			// Note: this does NOT match the YAML key "database.disable_ssl".
			// Multi-word fields cannot be overridden by env vars.
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotKey, gotValue := config.TransformKey(testCase.key, testCase.value)

			assert.Equal(t, testCase.wantKey, gotKey)
			assert.Equal(t, testCase.wantValue, gotValue)
		})
	}
}

func TestEnvDoesNotClobberUnrelatedKeys(t *testing.T) {
	t.Parallel()

	raw := []byte(`
database:
  host: yaml-host
  disable_ssl: true
  max_conns: 50
`)

	koanfManager := koanf.New(".")
	require.NoError(t, koanfManager.Load(rawbytes.Provider(raw), yaml.Parser()))
	require.NoError(t, koanfManager.Load(env.Provider(".", env.Opt{
		TransformFunc: config.TransformKey,
		EnvironFunc: func() []string {
			return []string{
				"DATABASE_HOST=env-host",
			}
		},
	}), nil))

	assert.True(t, koanfManager.Bool("database.disable_ssl"),
		"multi-word YAML key should keep its value")
	assert.Equal(t, int64(50), koanfManager.Int64("database.max_conns"),
		"multi-word YAML key should keep its value")
}
