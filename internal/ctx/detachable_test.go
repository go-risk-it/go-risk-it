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

func TestGameContext_Detach_PreservesMetadata(t *testing.T) {
	t.Parallel()

	gameCtx := newGameContext(context.Background(), "player-1", 99)

	detached, cancel := gameCtx.Detach(5 * time.Second)
	defer cancel()

	gc, ok := detached.(ctx.GameContext)
	require.True(t, ok, "detached context must be a GameContext")
	assert.Equal(t, int64(99), gc.GameID())
	assert.Equal(t, "player-1", gc.UserID())
}

func TestLobbyContext_Detach_PreservesMetadata(t *testing.T) {
	t.Parallel()

	lobbyCtx := newLobbyContext(context.Background(), "host-user", 42)

	detached, cancel := lobbyCtx.Detach(5 * time.Second)
	defer cancel()

	lc, ok := detached.(ctx.LobbyContext)
	require.True(t, ok, "detached context must be a LobbyContext")
	assert.Equal(t, int64(42), lc.LobbyID())
	assert.Equal(t, "host-user", lc.UserID())
}

func TestGameContext_Detach_DetachesFromParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	gameCtx := newGameContext(parent, "player-2", 7)

	detached, cancel := gameCtx.Detach(5 * time.Second)
	defer cancel()

	parentCancel()
	require.NoError(t, detached.Err(), "detached context must not inherit parent cancellation")
}

func TestLobbyContext_Detach_DetachesFromParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	lobbyCtx := newLobbyContext(parent, "host-2", 13)

	detached, cancel := lobbyCtx.Detach(5 * time.Second)
	defer cancel()

	parentCancel()
	require.NoError(t, detached.Err(), "detached context must not inherit parent cancellation")
}

func TestGameContext_Detach_SetsDeadline(t *testing.T) {
	t.Parallel()

	gameCtx := newGameContext(context.Background(), "player-3", 55)

	detached, cancel := gameCtx.Detach(1 * time.Millisecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)
	require.ErrorIs(t, detached.Err(), context.DeadlineExceeded)
}

func TestLobbyContext_Detach_SetsDeadline(t *testing.T) {
	t.Parallel()

	lobbyCtx := newLobbyContext(context.Background(), "host-3", 21)

	detached, cancel := lobbyCtx.Detach(1 * time.Millisecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)
	require.ErrorIs(t, detached.Err(), context.DeadlineExceeded)
}
