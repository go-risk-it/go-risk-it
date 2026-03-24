package runner

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/client"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
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
	want := orchestrator.DefaultTimeouts()

	assert.Equal(t, want.InitialStateWait, got.InitialStateWait)
	assert.Equal(t, want.UpdateWait, got.UpdateWait)
	assert.Equal(t, want.PhaseChangeWait, got.PhaseChangeWait)
	assert.Equal(t, want.PostMoveSettle, got.PostMoveSettle)
	assert.Equal(t, want.MaxConsecutiveErr, got.MaxConsecutiveErr)
}
