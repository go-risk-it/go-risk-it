package handlers_test

import (
	"sync"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateStore_BasicGetStore(t *testing.T) {
	t.Parallel()

	store := handlers.NewStateStore()
	state := &snapshot.CachedGameState{
		Turn: 1,
		PublicSnapshot: &snapshot.GameSnapshot{
			Game: snapshot.GameMeta{ID: 42, Turn: 1},
		},
	}

	store.Store(42, state)

	got := store.Get(42)
	require.NotNil(t, got)
	assert.Equal(t, state, got)
}

func TestStateStore_StaleWriteRejected(t *testing.T) {
	t.Parallel()

	store := handlers.NewStateStore()
	fresh := &snapshot.CachedGameState{
		Turn: 5,
		PublicSnapshot: &snapshot.GameSnapshot{
			Game: snapshot.GameMeta{ID: 42, Turn: 5},
		},
	}
	stale := &snapshot.CachedGameState{
		Turn: 3,
		PublicSnapshot: &snapshot.GameSnapshot{
			Game: snapshot.GameMeta{ID: 42, Turn: 3},
		},
	}

	store.Store(42, fresh)
	store.Store(42, stale) // should be silently ignored

	got := store.Get(42)
	require.NotNil(t, got)
	assert.Equal(t, int64(5), got.Turn)
	assert.Equal(t, fresh, got)
}

func TestStateStore_EqualTurnAccepted(t *testing.T) {
	t.Parallel()

	store := handlers.NewStateStore()
	first := &snapshot.CachedGameState{
		Turn: 5,
		PublicSnapshot: &snapshot.GameSnapshot{
			Game: snapshot.GameMeta{ID: 42, Turn: 5, WinnerUserID: ""},
		},
	}
	second := &snapshot.CachedGameState{
		Turn: 5,
		PublicSnapshot: &snapshot.GameSnapshot{
			Game: snapshot.GameMeta{ID: 42, Turn: 5, WinnerUserID: "player-1"},
		},
	}

	store.Store(42, first)
	store.Store(42, second) // same turn — last write wins

	got := store.Get(42)
	require.NotNil(t, got)
	assert.Equal(t, second, got)
	assert.Equal(t, "player-1", got.PublicSnapshot.Game.WinnerUserID)
}

func TestStateStore_Remove(t *testing.T) {
	t.Parallel()

	store := handlers.NewStateStore()
	store.Store(42, &snapshot.CachedGameState{Turn: 1})

	store.Remove(42)

	assert.Nil(t, store.Get(42))
}

func TestStateStore_RemoveIdempotent(t *testing.T) {
	t.Parallel()

	store := handlers.NewStateStore()

	// Removing a non-existent game must not panic.
	assert.NotPanics(t, func() {
		store.Remove(999)
	})
}

func TestStateStore_GetNonExistent(t *testing.T) {
	t.Parallel()

	store := handlers.NewStateStore()

	assert.Nil(t, store.Get(42))
}

func TestStateStore_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	store := handlers.NewStateStore()

	const numGames = 20
	const numOps = 100

	var wg sync.WaitGroup

	wg.Add(numGames)

	for gameID := range int64(numGames) {
		go func(id int64) {
			defer wg.Done()

			for i := range int64(numOps) {
				store.Store(id, &snapshot.CachedGameState{Turn: i})
				store.Get(id)
			}

			store.Remove(id)
		}(gameID)
	}

	wg.Wait()

	// All games were removed — store should be empty.
	for gameID := range int64(numGames) {
		assert.Nil(t, store.Get(gameID))
	}
}
