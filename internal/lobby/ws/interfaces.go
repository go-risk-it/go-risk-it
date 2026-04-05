package ws

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

// Writer broadcasts or sends messages to lobby WebSocket connections.
type Writer interface {
	Broadcast(ctx ctx.LobbyContext, message json.RawMessage)
	WriteMessage(ctx ctx.LobbyContext, message json.RawMessage)
}

// Gateway manages player WebSocket connection lifecycle for a lobby.
type Gateway interface {
	ConnectPlayer(ctx ctx.LobbyContext, connection *websocket.Conn)
}

// Manager composes all narrow lobby WebSocket interfaces.
// Consumers should depend on the narrowest interface they need.
//
// See game/ws/interfaces.go for the Generic Manager[C] evaluation (DROP).
// Lobby's 2-interface surface (Writer + Gateway, ~56 LOC) vs game's
// 4-interface surface (+ Presence + Lifecycle, ~139 LOC) is the structural
// asymmetry that makes a generic abstraction counterproductive.
type Manager interface {
	Writer
	Gateway
}

// Compile-time satisfaction checks.
var (
	_ Writer  = (*manager)(nil)
	_ Gateway = (*manager)(nil)
	_ Manager = (*manager)(nil)
)
