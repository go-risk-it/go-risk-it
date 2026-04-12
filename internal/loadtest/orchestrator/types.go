package orchestrator

import "time"

// GameResult holds stats from a completed game.
type GameResult struct {
	GameIndex  int
	Duration   time.Duration
	Moves      int
	Errors     int
	Winner     string
	TimedOut   bool
	Cancelled  bool
	FatalError error
}
