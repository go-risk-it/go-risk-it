package game

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
)

// StateStore caches the latest game snapshot keyed by game ID.
// Writes are guarded by a turn monotonicity check: a write is silently
// ignored if the incoming turn is strictly less than the stored turn.
// Thread-safe for concurrent reads and writes.
type StateStore interface {
	Get(gameID int64) *snapshot.CachedGameState
	Store(gameID int64, state *snapshot.CachedGameState)
	Remove(gameID int64)
}

// SnapshotReader reads game state as clean snapshot types.
// Implemented by the logic layer; consumed by publishers and controllers.
type SnapshotReader interface {
	GetPublicSnapshot(ctx ctx.GameContext) (*snapshot.GameSnapshot, error)
	GetAllPrivateSnapshots(ctx ctx.GameContext) (map[string]*snapshot.PlayerPrivate, error)
}

// StatePublisher sends a per-player view to a single connected client.
// Implemented by the web layer; consumed by event handlers after commits.
type StatePublisher interface {
	PublishState(ctx context.Context, playerUserID string, view *snapshot.PlayerView) error
}

// ScopeLifecycle manages the lifecycle of a scoped resource (e.g., WebSocket
// connection group for a game). Called on terminal events like game completion.
type ScopeLifecycle interface {
	RemoveScope(scopeID int64)
}
