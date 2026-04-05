package game

import (
	"log/slog"
	"time"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
)

type GameEvent interface {
	eventbus.Event
	GameID() int64
}

const (
	TypeMoveCompleted      = "move_completed"
	TypeGameCreated        = "game_created"
	TypeGameCreationFailed = "game_creation_failed"
	TypePlayerConnected    = "player_connected"
	TypeTurnEnded          = "turn_ended"
)

// MoveCompleted is the enriched post-commit event emitted by the orchestration
// pipeline. It carries full snapshot payloads and previous region state for
// downstream handlers (state broadcaster, headlines detector, lifecycle manager).
// No performer-specific types (AttackResult, CardsResult) cross the event boundary.
type MoveCompleted struct {
	gameID    int64
	userID    string
	timestamp time.Time

	ActionType  gameapi.GamePhaseType
	Turn        int64
	FromPhase   gameapi.GamePhaseType
	TargetPhase gameapi.GamePhaseType
	GameOver    bool

	PublicSnapshot   *snapshot.GameSnapshot
	PrivateSnapshots map[string]*snapshot.PlayerPrivate

	PreviousRegions []snapshot.RegionState
}

func NewMoveCompleted(
	gameID int64,
	userID string,
	timestamp time.Time,
	actionType gameapi.GamePhaseType,
	turn int64,
	fromPhase gameapi.GamePhaseType,
	targetPhase gameapi.GamePhaseType,
	gameOver bool,
	publicSnapshot *snapshot.GameSnapshot,
	privateSnapshots map[string]*snapshot.PlayerPrivate,
	previousRegions []snapshot.RegionState,
) *MoveCompleted {
	return &MoveCompleted{
		gameID:           gameID,
		userID:           userID,
		timestamp:        timestamp,
		ActionType:       actionType,
		Turn:             turn,
		FromPhase:        fromPhase,
		TargetPhase:      targetPhase,
		GameOver:         gameOver,
		PublicSnapshot:   publicSnapshot,
		PrivateSnapshots: privateSnapshots,
		PreviousRegions:  previousRegions,
	}
}

func (*MoveCompleted) EventType() string           { return TypeMoveCompleted }
func (e *MoveCompleted) GameID() int64             { return e.gameID }
func (e *MoveCompleted) EventTimestamp() time.Time { return e.timestamp }

func (e *MoveCompleted) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *MoveCompleted) ToRecord() map[string]any {
	return map[string]any{
		"event_type":   TypeMoveCompleted,
		"game_id":      e.gameID,
		"user_id":      e.userID,
		"timestamp":    e.timestamp.Format(time.RFC3339),
		"action_type":  string(e.ActionType),
		"turn":         e.Turn,
		"from_phase":   string(e.FromPhase),
		"target_phase": string(e.TargetPhase),
		"game_over":    e.GameOver,
	}
}

type GameCreated struct {
	gameID    int64
	lobbyID   int64
	timestamp time.Time

	NumPlayers int

	PublicSnapshot   *snapshot.GameSnapshot
	PrivateSnapshots map[string]*snapshot.PlayerPrivate
}

func NewGameCreated(
	gameID int64,
	lobbyID int64,
	timestamp time.Time,
	numPlayers int,
	publicSnapshot *snapshot.GameSnapshot,
	privateSnapshots map[string]*snapshot.PlayerPrivate,
) *GameCreated {
	return &GameCreated{
		gameID:           gameID,
		lobbyID:          lobbyID,
		timestamp:        timestamp,
		NumPlayers:       numPlayers,
		PublicSnapshot:   publicSnapshot,
		PrivateSnapshots: privateSnapshots,
	}
}

func (*GameCreated) EventType() string           { return TypeGameCreated }
func (e *GameCreated) GameID() int64             { return e.gameID }
func (e *GameCreated) LobbyID() int64            { return e.lobbyID }
func (e *GameCreated) EventTimestamp() time.Time { return e.timestamp }

func (e *GameCreated) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *GameCreated) ToRecord() map[string]any {
	record := map[string]any{
		"event_type":  TypeGameCreated,
		"game_id":     e.gameID,
		"timestamp":   e.timestamp.Format(time.RFC3339),
		"num_players": e.NumPlayers,
	}
	if e.lobbyID != 0 {
		record["lobby_id"] = e.lobbyID
	}

	return record
}

// GameCreationFailed is emitted when game creation fails. This is NOT a GameEvent
// (no gameID — the game was never created). It implements bus.Event directly.
type GameCreationFailed struct {
	lobbyID   int64
	timestamp time.Time
	Reason    string
}

func NewGameCreationFailed(
	lobbyID int64,
	timestamp time.Time,
	reason string,
) *GameCreationFailed {
	return &GameCreationFailed{
		lobbyID:   lobbyID,
		timestamp: timestamp,
		Reason:    reason,
	}
}

func (*GameCreationFailed) EventType() string           { return TypeGameCreationFailed }
func (e *GameCreationFailed) LobbyID() int64            { return e.lobbyID }
func (e *GameCreationFailed) EventTimestamp() time.Time { return e.timestamp }

func (e *GameCreationFailed) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("lobbyId", e.lobbyID)}
}

func (e *GameCreationFailed) ToRecord() map[string]any {
	return map[string]any{
		"event_type": TypeGameCreationFailed,
		"lobby_id":   e.lobbyID,
		"timestamp":  e.timestamp.Format(time.RFC3339),
		"reason":     e.Reason,
	}
}

type PlayerConnected struct {
	gameID    int64
	userID    string
	timestamp time.Time
}

func NewPlayerConnected(
	gameID int64,
	userID string,
	timestamp time.Time,
) *PlayerConnected {
	return &PlayerConnected{
		gameID:    gameID,
		userID:    userID,
		timestamp: timestamp,
	}
}

func (*PlayerConnected) EventType() string           { return TypePlayerConnected }
func (e *PlayerConnected) GameID() int64             { return e.gameID }
func (e *PlayerConnected) EventTimestamp() time.Time { return e.timestamp }

func (e *PlayerConnected) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *PlayerConnected) ToRecord() map[string]any {
	return map[string]any{
		"event_type": TypePlayerConnected,
		"game_id":    e.gameID,
		"user_id":    e.userID,
		"timestamp":  e.timestamp.Format(time.RFC3339),
	}
}

type TurnEnded struct {
	gameID    int64
	userID    string
	timestamp time.Time

	Turn int64
}

func NewTurnEnded(
	gameID int64,
	userID string,
	timestamp time.Time,
	turn int64,
) *TurnEnded {
	return &TurnEnded{
		gameID:    gameID,
		userID:    userID,
		timestamp: timestamp,
		Turn:      turn,
	}
}

func (*TurnEnded) EventType() string           { return TypeTurnEnded }
func (e *TurnEnded) GameID() int64             { return e.gameID }
func (e *TurnEnded) EventTimestamp() time.Time { return e.timestamp }

func (e *TurnEnded) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *TurnEnded) ToRecord() map[string]any {
	return map[string]any{
		"event_type":  TypeTurnEnded,
		"game_id":     e.gameID,
		"user_id":     e.userID,
		"timestamp":   e.timestamp.Format(time.RFC3339),
		"turn":        e.Turn,
		"action_type": "WAITING",
	}
}
