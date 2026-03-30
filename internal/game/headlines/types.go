package headlines

import (
	"log/slog"
	"time"
)

// Event type constants used as discriminators for bus routing and logging.
const (
	TypePlayerEliminated  = "player_eliminated"
	TypeContinentCaptured = "continent_captured"
	TypeContinentLost     = "continent_lost"
)

// PlayerEliminated is emitted when a player loses their last region.
type PlayerEliminated struct {
	gameID           int64
	eliminatedUserID string
	eliminatorUserID string
	timestamp        time.Time

	Turn int64
}

func NewPlayerEliminated(
	gameID int64,
	eliminatedUserID string,
	eliminatorUserID string,
	timestamp time.Time,
	turn int64,
) *PlayerEliminated {
	return &PlayerEliminated{
		gameID:           gameID,
		eliminatedUserID: eliminatedUserID,
		eliminatorUserID: eliminatorUserID,
		timestamp:        timestamp,
		Turn:             turn,
	}
}

func (*PlayerEliminated) EventType() string           { return TypePlayerEliminated }
func (e *PlayerEliminated) GameID() int64             { return e.gameID }
func (e *PlayerEliminated) EventTimestamp() time.Time { return e.timestamp }
func (e *PlayerEliminated) EliminatedUserID() string  { return e.eliminatedUserID }
func (e *PlayerEliminated) EliminatorUserID() string  { return e.eliminatorUserID }

func (e *PlayerEliminated) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *PlayerEliminated) ToRecord() map[string]any {
	return map[string]any{
		"event_type":         TypePlayerEliminated,
		"game_id":            e.gameID,
		"eliminated_user_id": e.eliminatedUserID,
		"eliminator_user_id": e.eliminatorUserID,
		"timestamp":          e.timestamp.Format(time.RFC3339),
		"turn":               e.Turn,
	}
}

// ContinentCaptured is emitted when a player gains control of all regions in a continent.
type ContinentCaptured struct {
	gameID    int64
	userID    string
	timestamp time.Time

	ContinentID string
	Turn        int64
}

func NewContinentCaptured(
	gameID int64,
	userID string,
	timestamp time.Time,
	continentID string,
	turn int64,
) *ContinentCaptured {
	return &ContinentCaptured{
		gameID:      gameID,
		userID:      userID,
		timestamp:   timestamp,
		ContinentID: continentID,
		Turn:        turn,
	}
}

func (*ContinentCaptured) EventType() string           { return TypeContinentCaptured }
func (e *ContinentCaptured) GameID() int64             { return e.gameID }
func (e *ContinentCaptured) EventTimestamp() time.Time { return e.timestamp }
func (e *ContinentCaptured) UserID() string            { return e.userID }

func (e *ContinentCaptured) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *ContinentCaptured) ToRecord() map[string]any {
	return map[string]any{
		"event_type":   TypeContinentCaptured,
		"game_id":      e.gameID,
		"user_id":      e.userID,
		"timestamp":    e.timestamp.Format(time.RFC3339),
		"continent_id": e.ContinentID,
		"turn":         e.Turn,
	}
}

// ContinentLost is emitted when a player loses complete control of a continent.
type ContinentLost struct {
	gameID    int64
	userID    string
	timestamp time.Time

	ContinentID string
	Turn        int64
}

func NewContinentLost(
	gameID int64,
	userID string,
	timestamp time.Time,
	continentID string,
	turn int64,
) *ContinentLost {
	return &ContinentLost{
		gameID:      gameID,
		userID:      userID,
		timestamp:   timestamp,
		ContinentID: continentID,
		Turn:        turn,
	}
}

func (*ContinentLost) EventType() string           { return TypeContinentLost }
func (e *ContinentLost) GameID() int64             { return e.gameID }
func (e *ContinentLost) EventTimestamp() time.Time { return e.timestamp }
func (e *ContinentLost) UserID() string            { return e.userID }

func (e *ContinentLost) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *ContinentLost) ToRecord() map[string]any {
	return map[string]any{
		"event_type":   TypeContinentLost,
		"game_id":      e.gameID,
		"user_id":      e.userID,
		"timestamp":    e.timestamp.Format(time.RFC3339),
		"continent_id": e.ContinentID,
		"turn":         e.Turn,
	}
}
