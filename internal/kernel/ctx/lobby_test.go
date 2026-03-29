package ctx_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestLobbyContext_DetachOnto_PreservesIDsAndDetachesFromParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	traceCtx := ctx.WithSpan(parent, noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "test-user")
	lobbyCtx := ctx.WithLobbyID(userCtx, 42)

	base, baseCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer baseCancel()

	detached := lobbyCtx.DetachOnto(base)

	lc, ok := detached.(ctx.LobbyContext)
	require.True(t, ok, "detached context must be a LobbyContext")
	require.Equal(t, int64(42), lc.LobbyID())
	require.Equal(t, "test-user", lc.UserID())

	// Cancelling the parent should NOT cancel the detached context.
	parentCancel()
	require.NoError(t, detached.Err(), "detached context must not inherit parent cancellation")
}

func TestLobbyContext_DetachOnto_InheritsBaseDeadline(t *testing.T) {
	t.Parallel()

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "test-user")
	lobbyCtx := ctx.WithLobbyID(userCtx, 7)

	base, baseCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer baseCancel()

	detached := lobbyCtx.DetachOnto(base)

	time.Sleep(10 * time.Millisecond)
	require.ErrorIs(t, detached.Err(), context.DeadlineExceeded)
}
