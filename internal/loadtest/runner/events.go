package runner

import (
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
)

// Event type constants.
const (
	EventGameStarted   EventType = "game_started"
	EventStateReceived EventType = "state_received"
	EventMoveDecided   EventType = "move_decided"
	EventTurnSkipped   EventType = "turn_skipped"
	EventMoveSucceeded EventType = "move_succeeded"
	EventMoveConflict  EventType = "move_conflict"
	EventMoveFailed    EventType = "move_failed"
	EventGameComplete  EventType = "game_complete"
)

// GameStartedEvent is emitted when a new game begins.
type GameStartedEvent struct {
	GameIndex  int
	NumPlayers int
}

func (GameStartedEvent) Type() EventType { return EventGameStarted }

// StateReceivedEvent carries a fresh game state snapshot.
type StateReceivedEvent struct {
	Snapshot  gamestate.ViewSnapshot
	Timestamp time.Time
}

func (StateReceivedEvent) Type() EventType { return EventStateReceived }

// MoveDecidedEvent carries a strategy decision ready for execution.
type MoveDecidedEvent struct {
	Action *player.Action
	UserID string
	Phase  metrics.Phase
}

func (MoveDecidedEvent) Type() EventType { return EventMoveDecided }

// TurnSkippedEvent indicates no action was taken this cycle.
type TurnSkippedEvent struct{}

func (TurnSkippedEvent) Type() EventType { return EventTurnSkipped }

// MoveSucceededEvent indicates a REST call completed successfully.
type MoveSucceededEvent struct {
	Action      *player.Action
	RESTLatency time.Duration
	RESTEndTime time.Time
}

func (MoveSucceededEvent) Type() EventType { return EventMoveSucceeded }

// MoveConflictEvent indicates a 409 conflict (stale view).
type MoveConflictEvent struct {
	Action *player.Action
}

func (MoveConflictEvent) Type() EventType { return EventMoveConflict }

// MoveFailedEvent indicates a REST call failed.
// ErrType carries the metrics error category so downstream handlers
// don't need to re-classify.
type MoveFailedEvent struct {
	Action  *player.Action
	Err     error
	Fatal   bool
	ErrType metrics.ErrorType
}

func (MoveFailedEvent) Type() EventType { return EventMoveFailed }

// GameCompleteEvent signals the game is over.
type GameCompleteEvent struct {
	Result GameResult
}

func (GameCompleteEvent) Type() EventType { return EventGameComplete }
