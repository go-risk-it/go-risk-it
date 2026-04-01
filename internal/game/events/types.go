package game

import (
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/cards"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
)

// GameEvent is the interface for game-scoped domain events. It embeds eventbus.Event
// and adds a GameID() accessor for game-specific scope identification.
type GameEvent interface {
	eventbus.Event
	GameID() int64
}

// Event type constants used as discriminators for bus routing and logging.
const (
	TypeMoveExecuted      = "move_executed"
	TypePhaseTransitioned = "phase_transitioned"
	TypeGameCompleted     = "game_completed"
	TypeGameCreated       = "game_created"
	TypePlayerConnected   = "player_connected"
)

// MoveExecuted is emitted after a move's transaction commits. It carries the complete
// outcome including action-specific results for attack and cards moves.
type MoveExecuted struct {
	gameID    int64
	userID    string
	timestamp time.Time

	ActionType  sqlc.GamePhaseType
	MoveLog     sqlc.GameMoveLog
	TargetPhase sqlc.GamePhaseType
	GameOver    bool
	Turn        int64

	AttackResult *attack.MoveResult
	CardsResult  *cards.MoveResult
}

func NewMoveExecuted(
	gameID int64,
	userID string,
	timestamp time.Time,
	actionType sqlc.GamePhaseType,
	moveLog sqlc.GameMoveLog,
	targetPhase sqlc.GamePhaseType,
	gameOver bool,
	turn int64,
	attackResult *attack.MoveResult,
	cardsResult *cards.MoveResult,
) *MoveExecuted {
	return &MoveExecuted{
		gameID:       gameID,
		userID:       userID,
		timestamp:    timestamp,
		ActionType:   actionType,
		MoveLog:      moveLog,
		TargetPhase:  targetPhase,
		GameOver:     gameOver,
		Turn:         turn,
		AttackResult: attackResult,
		CardsResult:  cardsResult,
	}
}

func (*MoveExecuted) EventType() string           { return TypeMoveExecuted }
func (e *MoveExecuted) GameID() int64             { return e.gameID }
func (e *MoveExecuted) EventTimestamp() time.Time { return e.timestamp }

func (e *MoveExecuted) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *MoveExecuted) ToRecord() map[string]any {
	record := map[string]any{
		"event_type":   TypeMoveExecuted,
		"game_id":      e.gameID,
		"user_id":      e.userID,
		"timestamp":    e.timestamp.Format(time.RFC3339),
		"action_type":  string(e.ActionType),
		"target_phase": string(e.TargetPhase),
		"game_over":    e.GameOver,
		"turn":         e.Turn,
		"move_log_id":  e.MoveLog.ID,
	}

	if e.AttackResult != nil {
		record["attacking_region_id"] = e.AttackResult.AttackingRegionID
		record["defending_region_id"] = e.AttackResult.DefendingRegionID
		record["conquering_troops"] = e.AttackResult.ConqueringTroops
	}

	if e.CardsResult != nil {
		record["extra_deployable_troops"] = e.CardsResult.ExtraDeployableTroops
		record["region_troop_grants"] = len(e.CardsResult.RegionTroopGrants)
	}

	return record
}

// PhaseTransitioned is emitted when the game phase changes.
type PhaseTransitioned struct {
	gameID    int64
	userID    string
	timestamp time.Time

	FromPhase sqlc.GamePhaseType
	ToPhase   sqlc.GamePhaseType
	Turn      int64
}

func NewPhaseTransitioned(
	gameID int64,
	userID string,
	timestamp time.Time,
	fromPhase sqlc.GamePhaseType,
	toPhase sqlc.GamePhaseType,
	turn int64,
) *PhaseTransitioned {
	return &PhaseTransitioned{
		gameID:    gameID,
		userID:    userID,
		timestamp: timestamp,
		FromPhase: fromPhase,
		ToPhase:   toPhase,
		Turn:      turn,
	}
}

func (*PhaseTransitioned) EventType() string           { return TypePhaseTransitioned }
func (e *PhaseTransitioned) GameID() int64             { return e.gameID }
func (e *PhaseTransitioned) EventTimestamp() time.Time { return e.timestamp }

func (e *PhaseTransitioned) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *PhaseTransitioned) ToRecord() map[string]any {
	return map[string]any{
		"event_type": TypePhaseTransitioned,
		"game_id":    e.gameID,
		"user_id":    e.userID,
		"timestamp":  e.timestamp.Format(time.RFC3339),
		"from_phase": string(e.FromPhase),
		"to_phase":   string(e.ToPhase),
		"turn":       e.Turn,
	}
}

// GameCompleted is emitted when a player wins by accomplishing their mission.
type GameCompleted struct {
	gameID       int64
	winnerUserID string
	timestamp    time.Time

	Turn int64
}

func NewGameCompleted(
	gameID int64,
	winnerUserID string,
	timestamp time.Time,
	turn int64,
) *GameCompleted {
	return &GameCompleted{
		gameID:       gameID,
		winnerUserID: winnerUserID,
		timestamp:    timestamp,
		Turn:         turn,
	}
}

func (*GameCompleted) EventType() string           { return TypeGameCompleted }
func (e *GameCompleted) GameID() int64             { return e.gameID }
func (e *GameCompleted) EventTimestamp() time.Time { return e.timestamp }

func (e *GameCompleted) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *GameCompleted) ToRecord() map[string]any {
	return map[string]any{
		"event_type":     TypeGameCompleted,
		"game_id":        e.gameID,
		"winner_user_id": e.winnerUserID,
		"timestamp":      e.timestamp.Format(time.RFC3339),
		"turn":           e.Turn,
	}
}

// GameCreated is emitted when a new game starts.
type GameCreated struct {
	gameID    int64
	timestamp time.Time

	NumPlayers int
}

func NewGameCreated(
	gameID int64,
	timestamp time.Time,
	numPlayers int,
) *GameCreated {
	return &GameCreated{
		gameID:     gameID,
		timestamp:  timestamp,
		NumPlayers: numPlayers,
	}
}

func (*GameCreated) EventType() string           { return TypeGameCreated }
func (e *GameCreated) GameID() int64             { return e.gameID }
func (e *GameCreated) EventTimestamp() time.Time { return e.timestamp }

func (e *GameCreated) ScopeAttrs() []slog.Attr {
	return []slog.Attr{slog.Int64("gameId", e.gameID)}
}

func (e *GameCreated) ToRecord() map[string]any {
	return map[string]any{
		"event_type":  TypeGameCreated,
		"game_id":     e.gameID,
		"timestamp":   e.timestamp.Format(time.RFC3339),
		"num_players": e.NumPlayers,
	}
}

// PlayerConnected is emitted when a player's WebSocket connects to a game.
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
