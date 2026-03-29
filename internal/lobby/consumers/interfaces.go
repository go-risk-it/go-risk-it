package consumers

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
)

// Writer sends messages to connected players. Used by event consumers that
// broadcast lobby state or write to individual players.
type Writer interface {
	Broadcast(ctx ctx.LobbyContext, message json.RawMessage)
	WriteMessage(ctx ctx.LobbyContext, message json.RawMessage)
}
