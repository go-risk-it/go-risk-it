package handlers_test

import (
	"testing"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeScopeLifecycle records RemoveScope calls.
type fakeScopeLifecycle struct {
	removed []int64
}

func (f *fakeScopeLifecycle) RemoveScope(scopeID int64) {
	f.removed = append(f.removed, scopeID)
}

var _ gameapi.ScopeLifecycle = (*fakeScopeLifecycle)(nil)

func moveCompletedGameOver(gameID int64, gameOver bool) *gameevt.MoveCompleted {
	return gameevt.NewMoveCompleted(
		gameID, testAttacker, fixedTime,
		gameapi.GamePhaseTypeATTACK,
		testTurn,
		gameapi.GamePhaseTypeATTACK,
		gameapi.GamePhaseTypeATTACK,
		gameOver, nil, nil, nil,
	)
}

func TestLifecycleManager_RemovesOnGameOver(t *testing.T) {
	t.Parallel()

	bus := newReentrantBus()
	lifecycle := &fakeScopeLifecycle{}
	store := handlers.NewStateStore()

	// Pre-populate state
	store.Store(testGameID, &snapshot.CachedGameState{Turn: 1})

	handlers.RegisterLifecycleManager(handlers.LifecycleManagerParams{
		Sub:            bus,
		ScopeLifecycle: lifecycle,
		StateStore:     store,
	})

	event := moveCompletedGameOver(testGameID, true)
	bus.Emit(gameCtx(testGameID), event)

	require.Len(t, lifecycle.removed, 1)
	assert.Equal(t, testGameID, lifecycle.removed[0])
	assert.Nil(t, store.Get(testGameID))
}

func TestLifecycleManager_IgnoresNonGameOver(t *testing.T) {
	t.Parallel()

	bus := newReentrantBus()
	lifecycle := &fakeScopeLifecycle{}
	store := handlers.NewStateStore()

	store.Store(testGameID, &snapshot.CachedGameState{Turn: 1})

	handlers.RegisterLifecycleManager(handlers.LifecycleManagerParams{
		Sub:            bus,
		ScopeLifecycle: lifecycle,
		StateStore:     store,
	})

	event := moveCompletedGameOver(testGameID, false)
	bus.Emit(gameCtx(testGameID), event)

	require.Empty(t, lifecycle.removed)
	assert.NotNil(t, store.Get(testGameID))
}
