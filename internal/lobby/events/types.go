package lobby

import (
	"time"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
)

// LobbyEvent is the interface for lobby-scoped domain events. It embeds eventbus.Event
// and adds a LobbyID() accessor for lobby-specific scope identification.
type LobbyEvent interface {
	eventbus.Event
	LobbyID() int64
}

// Event type constants used as discriminators for bus routing and logging.
const (
	TypeLobbyStateChanged    = "lobby_state_changed"
	TypeLobbyPlayerConnected = "lobby_player_connected"
)

// LobbyStateChanged is emitted when the lobby state changes (e.g., player joins/leaves,
// settings updated).
type LobbyStateChanged struct {
	lobbyID   int64
	userID    string
	timestamp time.Time
}

func NewLobbyStateChanged(lobbyID int64, userID string) *LobbyStateChanged {
	return &LobbyStateChanged{
		lobbyID:   lobbyID,
		userID:    userID,
		timestamp: time.Now(),
	}
}

func (*LobbyStateChanged) EventType() string           { return TypeLobbyStateChanged }
func (e *LobbyStateChanged) LobbyID() int64            { return e.lobbyID }
func (e *LobbyStateChanged) UserID() string            { return e.userID }
func (e *LobbyStateChanged) EventTimestamp() time.Time { return e.timestamp }

func (e *LobbyStateChanged) ToRecord() map[string]any {
	return map[string]any{
		"event_type": TypeLobbyStateChanged,
		"lobby_id":   e.lobbyID,
		"user_id":    e.userID,
		"timestamp":  e.timestamp.Format(time.RFC3339),
	}
}

// LobbyPlayerConnected is emitted when a player's WebSocket connects to a lobby.
type LobbyPlayerConnected struct {
	lobbyID   int64
	userID    string
	timestamp time.Time
}

func NewLobbyPlayerConnected(lobbyID int64, userID string) *LobbyPlayerConnected {
	return &LobbyPlayerConnected{
		lobbyID:   lobbyID,
		userID:    userID,
		timestamp: time.Now(),
	}
}

func (*LobbyPlayerConnected) EventType() string           { return TypeLobbyPlayerConnected }
func (e *LobbyPlayerConnected) LobbyID() int64            { return e.lobbyID }
func (e *LobbyPlayerConnected) UserID() string            { return e.userID }
func (e *LobbyPlayerConnected) EventTimestamp() time.Time { return e.timestamp }

func (e *LobbyPlayerConnected) ToRecord() map[string]any {
	return map[string]any{
		"event_type": TypeLobbyPlayerConnected,
		"lobby_id":   e.lobbyID,
		"user_id":    e.userID,
		"timestamp":  e.timestamp.Format(time.RFC3339),
	}
}
