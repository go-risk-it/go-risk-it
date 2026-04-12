package ctx_test

import (
	"context"
	"testing"
	"time"

	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestLobbyContext_Rebase_PreservesIDsAndDetachesFromParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	traceCtx := kernelctx.WithSpan(parent, noop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "test-user")
	lobbyCtx := ctx.WithLobbyID(userCtx, 42)

	base, baseCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer baseCancel()

	rebased := lobbyCtx.Rebase(base)

	lc, ok := rebased.(ctx.LobbyContext)
	require.True(t, ok, "rebased context must be a LobbyContext")
	require.Equal(t, int64(42), lc.LobbyID())
	require.Equal(t, "test-user", lc.UserID())

	// Cancelling the parent should NOT cancel the rebased context.
	parentCancel()
	require.NoError(t, rebased.Err(), "rebased context must not inherit parent cancellation")
}

func TestLobbyContext_Rebase_InheritsBaseDeadline(t *testing.T) {
	t.Parallel()

	traceCtx := kernelctx.WithSpan(context.Background(), noop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "test-user")
	lobbyCtx := ctx.WithLobbyID(userCtx, 7)

	base, baseCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer baseCancel()

	rebased := lobbyCtx.Rebase(base)

	time.Sleep(10 * time.Millisecond)
	require.ErrorIs(t, rebased.Err(), context.DeadlineExceeded)
}

func TestLobbyContext_ScopeID(t *testing.T) {
	t.Parallel()

	traceCtx := kernelctx.WithSpan(context.Background(), noop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "test-user")
	lobbyCtx := ctx.WithLobbyID(userCtx, 42)

	require.Equal(t, lobbyCtx.LobbyID(), lobbyCtx.ScopeID(),
		"ScopeID must equal LobbyID")
}
