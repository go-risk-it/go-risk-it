package orchestrator

import "time"

// Timeouts holds all configurable timing parameters for the game loop.
type Timeouts struct {
	InitialStateWait  time.Duration // wait for first WS state after connect
	UpdateWait        time.Duration // wait for state update after move
	PhaseChangeWait   time.Duration // wait for phase change on 409
	PostMoveSettle    time.Duration // settle time after WS update
	MaxConsecutiveErr int           // consecutive errors before fatal
}

// DefaultTimeouts returns sensible defaults for all timing parameters.
func DefaultTimeouts() Timeouts {
	return Timeouts{
		InitialStateWait:  1 * time.Second,
		UpdateWait:        3 * time.Second,
		PhaseChangeWait:   3 * time.Second,
		PostMoveSettle:    50 * time.Millisecond,
		MaxConsecutiveErr: 20,
	}
}

// GameResult holds stats from a completed game.
type GameResult struct {
	GameIndex  int
	Duration   time.Duration
	Moves      int
	Errors     int
	Winner     string
	TimedOut   bool
	FatalError error
}
