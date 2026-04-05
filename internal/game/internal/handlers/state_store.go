package handlers

import (
	"sync"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
)

// stateStore is an in-memory cache of game snapshots keyed by game ID.
// Writes are guarded by turn monotonicity: a Store call is silently
// ignored when the incoming turn is strictly less than the stored turn.
type stateStore struct {
	mu    sync.RWMutex
	cache map[int64]*snapshot.CachedGameState
}

var _ gameapi.StateStore = (*stateStore)(nil)

// NewStateStore creates a ready-to-use StateStore.
func NewStateStore() gameapi.StateStore {
	return &stateStore{
		cache: make(map[int64]*snapshot.CachedGameState),
	}
}

// Get returns the cached state for gameID, or nil if not present.
func (s *stateStore) Get(gameID int64) *snapshot.CachedGameState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.cache[gameID]
}

// Store caches the given state for gameID. The write is silently
// ignored if the incoming state's Turn is strictly less than the
// already-stored Turn (stale-write guard). Equal turns are accepted
// (last-write-wins at the same turn).
func (s *stateStore) Store(gameID int64, state *snapshot.CachedGameState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.cache[gameID]; ok && state.Turn < existing.Turn {
		return
	}

	s.cache[gameID] = state
}

// Remove deletes the cached state for gameID. No-op if the game is
// not in the cache.
func (s *stateStore) Remove(gameID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.cache, gameID)
}
