package config //nolint:testpackage // whitebox tests for unexported readFromEnv and getEnv

import (
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyTransform(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envValue string
		wantKey  string
	}{
		{
			name:     "single hierarchy",
			envVar:   "DATABASE_HOST",
			envValue: "myhost",
			wantKey:  "database.host",
		},
		{
			name:     "two levels",
			envVar:   "JWT_SECRET",
			envValue: "mysecret",
			wantKey:  "jwt.secret",
		},
		{
			name:     "boolean",
			envVar:   "OTEL_ENABLED",
			envValue: "true",
			wantKey:  "otel.enabled",
		},
		{
			name:     "multi-word field becomes extra dot",
			envVar:   "DATABASE_DISABLE_SSL",
			envValue: "true",
			wantKey:  "database.disable.ssl",
			// Note: this does NOT match the YAML key "database.disable_ssl".
			// Multi-word fields cannot be overridden by env vars. By design,
			// secrets come from env vars (single-word keys) and structure
			// comes from YAML.
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(testCase.envVar, testCase.envValue)

			koanfManager := koanf.New(".")
			require.NoError(t, readFromEnv(koanfManager))

			assert.Equal(t, testCase.envValue, koanfManager.String(testCase.wantKey))
		})
	}
}

func TestReadFromEnv_OverridesExistingKeys(t *testing.T) {
	raw := []byte(`
database:
  host: yaml-host
  port: 5432
`)

	t.Setenv("DATABASE_HOST", "env-host")

	koanfManager := koanf.New(".")
	require.NoError(t, koanfManager.Load(rawbytes.Provider(raw), yaml.Parser()))
	require.NoError(t, readFromEnv(koanfManager))

	assert.Equal(t, "env-host", koanfManager.String("database.host"),
		"env var should override YAML value")
	assert.Equal(t, int64(5432), koanfManager.Int64("database.port"),
		"YAML-only key should be preserved")
}

func TestReadFromEnv_DoesNotClobberUnrelatedKeys(t *testing.T) {
	raw := []byte(`
database:
  host: yaml-host
  disable_ssl: true
  max_conns: 50
`)

	t.Setenv("DATABASE_HOST", "env-host")

	koanfManager := koanf.New(".")
	require.NoError(t, koanfManager.Load(rawbytes.Provider(raw), yaml.Parser()))
	require.NoError(t, readFromEnv(koanfManager))

	assert.True(t, koanfManager.Bool("database.disable_ssl"),
		"multi-word YAML key should keep its value")
	assert.Equal(t, int64(50), koanfManager.Int64("database.max_conns"),
		"multi-word YAML key should keep its value")
}

func TestGetEnv_Default(t *testing.T) {
	t.Parallel()

	// ENVIRONMENT is not set — getEnv should return the default.
	assert.Equal(t, "component-test", getEnv())
}

func TestGetEnv_Custom(t *testing.T) {
	t.Setenv("ENVIRONMENT", "prod")

	assert.Equal(t, "prod", getEnv())
}
