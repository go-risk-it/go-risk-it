package consumers

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
)

// Writer sends messages to connected players. Used by event consumers that
// broadcast game state or write to individual players.
type Writer interface {
	Broadcast(ctx ctx.GameContext, message json.RawMessage)
	WriteMessage(ctx ctx.GameContext, message json.RawMessage)
}

// Presence exposes the set of currently connected players for a game. Used by
// converters that need to annotate state with connection status.
type Presence interface {
	GetConnectedPlayers(ctx ctx.GameContext) []string
}

// Lifecycle manages the lifetime of a game's connection tracking. Used by the
// GameCompleted consumer to clean up after a game ends.
type Lifecycle interface {
	RemoveGame(ctx ctx.GameContext)
}
