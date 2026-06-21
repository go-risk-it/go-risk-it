package orchestration

import (
	"sync"
)

// GameLocks provides per-game mutual exclusion. Each game ID gets its own
// mutex, lazily created on first access. This allows concurrent moves on
// different games while serializing moves on the same game outside the DB
// transaction (e.g., for cache reads that must see a consistent snapshot).
type GameLocks struct {
	locks sync.Map // map[int64]*sync.Mutex
}

// NewGameLocks creates a new GameLocks instance.
func NewGameLocks() *GameLocks {
	return &GameLocks{}
}

// Lock acquires the mutex for the given game ID, blocking if another
// goroutine already holds it. The returned function releases the lock.
// The release function is idempotent: calling it more than once is safe.
func (gl *GameLocks) Lock(gameID int64) func() {
	actual, _ := gl.locks.LoadOrStore(gameID, &sync.Mutex{})

	gameMutex, ok := actual.(*sync.Mutex)
	if !ok {
		panic("GameLocks: unexpected type in sync.Map")
	}

	gameMutex.Lock()

	var once sync.Once

	return func() {
		once.Do(gameMutex.Unlock)
	}
}
