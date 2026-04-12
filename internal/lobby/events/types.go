package lobby

import (
	"log/slog"
	"time"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/snapshot"
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
	TypeCreateGameRequested  = "create_game_requested"
)

// LobbyStateChanged is emitted when the lobby state changes (e.g., player joins/leaves,
// settings updated). Carries the full LobbySnapshot for zero-query broadcast.
type LobbyStateChanged struct {
	lobbyID   int64
	userID    string
	timestamp time.Time

	Snapshot *snapshot.LobbySnapshot
}

func NewLobbyStateChanged(
	lobbyID int64,
	userID string,
	snap *snapshot.LobbySnapshot,
) *LobbyStateChanged {
	return &LobbyStateChanged{
		lobbyID:   lobbyID,
		userID:    userID,
		timestamp: time.Now(),
		Snapshot:  snap,
	}
}

func (*LobbyStateChanged) EventType() string           { return TypeLobbyStateChanged }
func (e *LobbyStateChanged) LobbyID() int64            { return e.lobbyID }
func (e *LobbyStateChanged) UserID() string            { return e.userID }
func (e *LobbyStateChanged) EventTimestamp() time.Time { return e.timestamp }

func (e *LobbyStateChanged) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("lobbyId", e.lobbyID)}
}

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

func (e *LobbyPlayerConnected) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("lobbyId", e.lobbyID)}
}

func (e *LobbyPlayerConnected) ToRecord() map[string]any {
	return map[string]any{
		"event_type": TypeLobbyPlayerConnected,
		"lobby_id":   e.lobbyID,
		"user_id":    e.userID,
		"timestamp":  e.timestamp.Format(time.RFC3339),
	}
}

// LobbyPlayer represents a player in a lobby for cross-module game creation requests.
type LobbyPlayer struct {
	UserID string
	Name   string
}

// CreateGameRequested is emitted when a lobby owner starts a game. This is the cross-module
// signal that the game module consumes to create the actual game.
type CreateGameRequested struct {
	lobbyID   int64
	userID    string
	timestamp time.Time
	Players   []LobbyPlayer
}

func NewCreateGameRequested(
	lobbyID int64,
	userID string,
	timestamp time.Time,
	players []LobbyPlayer,
) *CreateGameRequested {
	return &CreateGameRequested{
		lobbyID:   lobbyID,
		userID:    userID,
		timestamp: timestamp,
		Players:   players,
	}
}

func (*CreateGameRequested) EventType() string           { return TypeCreateGameRequested }
func (e *CreateGameRequested) LobbyID() int64            { return e.lobbyID }
func (e *CreateGameRequested) UserID() string            { return e.userID }
func (e *CreateGameRequested) EventTimestamp() time.Time { return e.timestamp }

func (e *CreateGameRequested) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("lobbyId", e.lobbyID)}
}

func (e *CreateGameRequested) ToRecord() map[string]any {
	return map[string]any{
		"event_type":  TypeCreateGameRequested,
		"lobby_id":    e.lobbyID,
		"user_id":     e.userID,
		"timestamp":   e.timestamp.Format(time.RFC3339),
		"num_players": len(e.Players),
	}
}
