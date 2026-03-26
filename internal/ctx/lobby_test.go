package ctx_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestDetachLobbyContextWithTimeout_PreservesLobbyID(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	traceCtx := ctx.WithSpan(parent, noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "test-user")
	lobbyCtx := ctx.WithLobbyID(userCtx, 42)

	detached, cancel := ctx.DetachLobbyContextWithTimeout(lobbyCtx, 5*time.Second)
	defer cancel()

	assert.Equal(t, int64(42), detached.LobbyID())
	assert.Equal(t, "test-user", detached.UserID())

	// Cancelling the parent should NOT cancel the detached context.
	parentCancel()
	require.NoError(t, detached.Err(), "detached context must not inherit parent cancellation")
}

func TestDetachLobbyContextWithTimeout_TimesOut(t *testing.T) {
	t.Parallel()

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "test-user")
	lobbyCtx := ctx.WithLobbyID(userCtx, 7)

	detached, cancel := ctx.DetachLobbyContextWithTimeout(lobbyCtx, 1*time.Millisecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)
	require.ErrorIs(t, detached.Err(), context.DeadlineExceeded)
}
