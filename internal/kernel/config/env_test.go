package config //nolint:testpackage // whitebox tests for unexported getEnv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv_Default(t *testing.T) {
	t.Parallel()

	// ENVIRONMENT is not set — getEnv should return the default.
	assert.Equal(t, "component-test", getEnv())
}

func TestGetEnv_Custom(t *testing.T) {
	t.Setenv("ENVIRONMENT", "prod")

	assert.Equal(t, "prod", getEnv())
}
