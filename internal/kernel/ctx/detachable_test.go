package ctx_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func newGameContext(parent context.Context, userID string, gameID int64) ctx.GameContext {
	traceCtx := ctx.WithSpan(parent, noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, userID)

	return ctx.WithGameID(userCtx, gameID)
}

func newLobbyContext(parent context.Context, userID string, lobbyID int64) ctx.LobbyContext {
	traceCtx := ctx.WithSpan(parent, noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, userID)

	return ctx.WithLobbyID(userCtx, lobbyID)
}

func TestGameContext_DetachOnto_PreservesMetadata(t *testing.T) {
	t.Parallel()

	gameCtx := newGameContext(context.Background(), "player-1", 99)

	detached := gameCtx.DetachOnto(context.Background())

	gc, ok := detached.(ctx.GameContext)
	require.True(t, ok, "detached context must be a GameContext")
	require.Equal(t, int64(99), gc.GameID())
	require.Equal(t, "player-1", gc.UserID())
}

func TestLobbyContext_DetachOnto_PreservesMetadata(t *testing.T) {
	t.Parallel()

	lobbyCtx := newLobbyContext(context.Background(), "host-user", 42)

	detached := lobbyCtx.DetachOnto(context.Background())

	lc, ok := detached.(ctx.LobbyContext)
	require.True(t, ok, "detached context must be a LobbyContext")
	require.Equal(t, int64(42), lc.LobbyID())
	require.Equal(t, "host-user", lc.UserID())
}

func TestGameContext_DetachOnto_DetachesFromParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	gameCtx := newGameContext(parent, "player-2", 7)

	base, baseCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer baseCancel()

	detached := gameCtx.DetachOnto(base)

	parentCancel()
	require.NoError(t, detached.Err(), "detached context must not inherit parent cancellation")
}

func TestLobbyContext_DetachOnto_DetachesFromParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	lobbyCtx := newLobbyContext(parent, "host-2", 13)

	base, baseCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer baseCancel()

	detached := lobbyCtx.DetachOnto(base)

	parentCancel()
	require.NoError(t, detached.Err(), "detached context must not inherit parent cancellation")
}

func TestGameContext_DetachOnto_PropagatesBaseCancellation(t *testing.T) {
	t.Parallel()

	base, baseCancel := context.WithCancel(context.Background())
	gameCtx := newGameContext(context.Background(), "player-4", 33)

	detached := gameCtx.DetachOnto(base)

	require.NoError(t, detached.Err(), "detached context must not be cancelled before base")

	baseCancel()
	require.Error(t, detached.Err(), "detached context must propagate base cancellation")
}

func TestLobbyContext_DetachOnto_PropagatesBaseCancellation(t *testing.T) {
	t.Parallel()

	base, baseCancel := context.WithCancel(context.Background())
	lobbyCtx := newLobbyContext(context.Background(), "host-4", 77)

	detached := lobbyCtx.DetachOnto(base)

	require.NoError(t, detached.Err(), "detached context must not be cancelled before base")

	baseCancel()
	require.Error(t, detached.Err(), "detached context must propagate base cancellation")
}

func TestGameContext_DetachOnto_Deadline(t *testing.T) {
	t.Parallel()

	gameCtx := newGameContext(context.Background(), "player-3", 55)

	base, baseCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer baseCancel()

	detached := gameCtx.DetachOnto(base)

	time.Sleep(10 * time.Millisecond)
	require.ErrorIs(t, detached.Err(), context.DeadlineExceeded)
}

func TestLobbyContext_DetachOnto_Deadline(t *testing.T) {
	t.Parallel()

	lobbyCtx := newLobbyContext(context.Background(), "host-3", 21)

	base, baseCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer baseCancel()

	detached := lobbyCtx.DetachOnto(base)

	time.Sleep(10 * time.Millisecond)
	require.ErrorIs(t, detached.Err(), context.DeadlineExceeded)
}
