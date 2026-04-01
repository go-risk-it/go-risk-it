package baseline_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
	"github.com/stretchr/testify/assert"
)

func TestCaptureEnvironment_RuntimeFields(t *testing.T) {
	t.Parallel()

	env := baseline.CaptureEnvironment()

	assert.NotEmpty(t, env.GoVersion, "GoVersion should be populated")
	assert.NotEmpty(t, env.GOOS, "GOOS should be populated")
	assert.NotEmpty(t, env.GOARCH, "GOARCH should be populated")
	assert.Positive(t, env.GOMAXPROCS, "GOMAXPROCS should be positive")
	assert.Positive(t, env.NumCPU, "NumCPU should be positive")
}
