package ws

import (
	upgradablerwmutex "github.com/go-risk-it/go-risk-it/internal/kernel/upgradablerw_mutex"
)

// ScopeMap is a concurrent registry mapping scope identifiers (game IDs,
// lobby IDs) to their [PlayerConnections]. It encapsulates the
// [upgradablerwmutex.UpgradableRWMutex] double-check locking pattern used
// by both game and lobby WebSocket managers.
//
// Three methods map to existing manager access patterns:
//
//	Get         — read-only lookup, returns nil on miss (broadcast/write path)
//	GetOrCreate — double-check locking creation (connect path)
//	Remove      — exclusive deletion, returns value for metric bookkeeping
//
// # Mapping to game manager methods
//
//	manager.getPlayerConnections   → ScopeMap.Get
//	manager.getOrCreatePlayerConnections → ScopeMap.GetOrCreate
//	manager.RemoveGame             → ScopeMap.Remove
type ScopeMap[K comparable] struct {
	mu      upgradablerwmutex.UpgradableRWMutex
	entries map[K]*PlayerConnections
}

// NewScopeMap creates an empty [ScopeMap] ready for use.
func NewScopeMap[K comparable]() *ScopeMap[K] {
	return &ScopeMap[K]{
		entries: make(map[K]*PlayerConnections),
	}
}

// Get returns the [PlayerConnections] for the given key, or nil if the key is
// not tracked. This is the read-only path: it never creates a new entry, so
// callers must handle a nil return (typically by returning early/no-op).
//
// Uses UpgradableRLock (not RLock) for race-detector compatibility with
// [ScopeMap.Remove]'s exclusive Lock — UpgradableRWMutex's RLock lacks
// race annotations.
func (s *ScopeMap[K]) Get(key K) *PlayerConnections {
	s.mu.UpgradableRLock()
	defer s.mu.UpgradableRUnlock()

	return s.entries[key]
}

// GetOrCreate returns the [PlayerConnections] for the given key, creating one
// via factory if the key is not tracked. Uses double-check locking: acquires a
// read lock first, then upgrades to a write lock only when needed, re-checking
// after the upgrade to avoid duplicate creation.
//
// The factory function must not call back into the ScopeMap (deadlock).
func (s *ScopeMap[K]) GetOrCreate(key K, factory func() *PlayerConnections) *PlayerConnections {
	s.mu.UpgradableRLock()
	defer s.mu.UpgradableRUnlock()

	connections, ok := s.entries[key]
	if ok {
		return connections
	}

	s.mu.UpgradeWLock()

	// Re-check after acquiring write lock — another goroutine may have inserted.
	if existing, exists := s.entries[key]; exists {
		return existing
	}

	connections = factory()
	s.entries[key] = connections

	return connections
}

// Remove deletes the [PlayerConnections] for the given key under an exclusive
// lock. Returns the removed entry and true, or (nil, false) if the key was not
// tracked. The caller is responsible for metric bookkeeping (e.g. decrementing
// ActiveConnections by the removed entry's PlayerCount).
func (s *ScopeMap[K]) Remove(key K) (*PlayerConnections, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	connections, ok := s.entries[key]
	if !ok {
		return nil, false
	}

	delete(s.entries, key)

	return connections, true
}
