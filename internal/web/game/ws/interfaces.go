package ws

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

// Writer sends messages to connected players. Used by event consumers that
// broadcast game state or write to individual players.
type Writer interface {
	Broadcast(ctx ctx.GameContext, message json.RawMessage)
	WriteMessage(ctx ctx.GameContext, message json.RawMessage)
}

// Presence exposes the set of currently connected players for a game. Used by
// controllers and converters that need to annotate state with connection status.
type Presence interface {
	GetConnectedPlayers(ctx ctx.GameContext) []string
}

// Lifecycle manages the lifetime of a game's connection tracking. Used by the
// GameCompleted consumer to clean up after a game ends.
type Lifecycle interface {
	RemoveGame(ctx ctx.GameContext)
}

// Gateway handles new WebSocket connection establishment. Used by the REST
// handler that upgrades HTTP connections to WebSocket.
type Gateway interface {
	ConnectPlayer(ctx ctx.GameContext, connection *websocket.Conn)
}

// Compile-time satisfaction checks — the concrete manager must implement all
// narrow interfaces.
var (
	_ Writer    = (*manager)(nil)
	_ Presence  = (*manager)(nil)
	_ Lifecycle = (*manager)(nil)
	_ Gateway   = (*manager)(nil)
)

// Manager is the aggregate interface embedding all narrow WS capabilities.
// Existing callers that need the full surface continue to depend on Manager;
// new callers should depend on the narrowest interface they need (Writer,
// Presence, Lifecycle, or Gateway).
type Manager interface {
	Writer
	Presence
	Lifecycle
	Gateway
}

var _ Manager = (*manager)(nil)
