package ctx_test

import (
	"context"
	"testing"
	"time"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	lobbyclx "github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func newUserContext(parent context.Context, userID string) ctx.UserContext {
	traceCtx := ctx.WithSpan(parent, noop.Span{})

	return ctx.WithUserID(traceCtx, userID)
}

func newGameContext(parent context.Context, userID string, gameID int64) gamectx.GameContext {
	traceCtx := ctx.WithSpan(parent, noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, userID)

	return gamectx.WithGameID(userCtx, gameID)
}

func newLobbyContext(parent context.Context, userID string, lobbyID int64) lobbyclx.LobbyContext {
	traceCtx := ctx.WithSpan(parent, noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, userID)

	return lobbyclx.WithLobbyID(userCtx, lobbyID)
}

func TestGameContext_Rebase_PreservesMetadata(t *testing.T) {
	t.Parallel()

	gameCtx := newGameContext(context.Background(), "player-1", 99)

	rebased := gameCtx.Rebase(context.Background())

	gc, ok := rebased.(gamectx.GameContext)
	require.True(t, ok, "rebased context must be a GameContext")
	require.Equal(t, int64(99), gc.GameID())
	require.Equal(t, "player-1", gc.UserID())
}

func TestLobbyContext_Rebase_PreservesMetadata(t *testing.T) {
	t.Parallel()

	lobbyCtx := newLobbyContext(context.Background(), "host-user", 42)

	rebased := lobbyCtx.Rebase(context.Background())

	lc, ok := rebased.(lobbyclx.LobbyContext)
	require.True(t, ok, "rebased context must be a LobbyContext")
	require.Equal(t, int64(42), lc.LobbyID())
	require.Equal(t, "host-user", lc.UserID())
}

func TestUserContext_Rebase_PreservesMetadata(t *testing.T) {
	t.Parallel()

	userCtx := newUserContext(context.Background(), "user-abc")

	//nolint:forcetypeassert // test verifies the Rebaseable contract
	rebased := userCtx.(ctx.Rebaseable).Rebase(context.Background())

	uc, ok := rebased.(ctx.UserContext)
	require.True(t, ok, "rebased context must be a UserContext")
	require.Equal(t, "user-abc", uc.UserID())
}

func TestGameContext_Rebase_DetachesFromParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	gameCtx := newGameContext(parent, "player-2", 7)

	base, baseCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer baseCancel()

	rebased := gameCtx.Rebase(base)

	parentCancel()
	require.NoError(t, rebased.Err(), "rebased context must not inherit parent cancellation")
}

func TestLobbyContext_Rebase_DetachesFromParent(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	lobbyCtx := newLobbyContext(parent, "host-2", 13)

	base, baseCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer baseCancel()

	rebased := lobbyCtx.Rebase(base)

	parentCancel()
	require.NoError(t, rebased.Err(), "rebased context must not inherit parent cancellation")
}

func TestGameContext_Rebase_PropagatesBaseCancellation(t *testing.T) {
	t.Parallel()

	base, baseCancel := context.WithCancel(context.Background())
	gameCtx := newGameContext(context.Background(), "player-4", 33)

	rebased := gameCtx.Rebase(base)

	require.NoError(t, rebased.Err(), "rebased context must not be cancelled before base")

	baseCancel()
	require.Error(t, rebased.Err(), "rebased context must propagate base cancellation")
}

func TestLobbyContext_Rebase_PropagatesBaseCancellation(t *testing.T) {
	t.Parallel()

	base, baseCancel := context.WithCancel(context.Background())
	lobbyCtx := newLobbyContext(context.Background(), "host-4", 77)

	rebased := lobbyCtx.Rebase(base)

	require.NoError(t, rebased.Err(), "rebased context must not be cancelled before base")

	baseCancel()
	require.Error(t, rebased.Err(), "rebased context must propagate base cancellation")
}

func TestGameContext_Rebase_InheritsBaseDeadline(t *testing.T) {
	t.Parallel()

	gameCtx := newGameContext(context.Background(), "player-3", 55)

	base, baseCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer baseCancel()

	rebased := gameCtx.Rebase(base)

	time.Sleep(10 * time.Millisecond)
	require.ErrorIs(t, rebased.Err(), context.DeadlineExceeded)
}

func TestLobbyContext_Rebase_InheritsBaseDeadline(t *testing.T) {
	t.Parallel()

	lobbyCtx := newLobbyContext(context.Background(), "host-3", 21)

	base, baseCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer baseCancel()

	rebased := lobbyCtx.Rebase(base)

	time.Sleep(10 * time.Millisecond)
	require.ErrorIs(t, rebased.Err(), context.DeadlineExceeded)
}
