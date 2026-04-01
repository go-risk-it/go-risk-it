package runner //nolint:testpackage // whitebox tests access unexported helpers

import (
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/stretchr/testify/assert"
)

// Compile-time interface satisfaction checks.
var (
	_ RESTClient = (*client.REST)(nil)
	_ WSClient   = (*client.WS)(nil)
	_ AuthClient = (*client.Auth)(nil)
)

func TestDefaultTimeouts(t *testing.T) {
	t.Parallel()

	got := DefaultTimeouts()

	assert.Equal(t, 1*time.Second, got.InitialStateWait)
	assert.Equal(t, 3*time.Second, got.UpdateWait)
	assert.Equal(t, 3*time.Second, got.PhaseChangeWait)
	assert.Equal(t, 50*time.Millisecond, got.PostMoveSettle)
	assert.Equal(t, 20, got.MaxConsecutiveErr)
}
