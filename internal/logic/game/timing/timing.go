// Package timing tracks game start times for duration metric calculation.
package timing

import (
	"sync"
	"time"

	"go.uber.org/fx"
)

// Module provides GameTiming as an fx dependency.
var Module = fx.Options(
	fx.Provide(NewGameTiming),
)

// GameTiming tracks when games start so that game duration can be
// computed when the game finishes. Safe for concurrent use.
type GameTiming struct {
	startTimes sync.Map
}

// NewGameTiming creates a new GameTiming tracker.
func NewGameTiming() *GameTiming {
	return &GameTiming{}
}

// RecordStart stores the start time for a game.
func (g *GameTiming) RecordStart(gameID int64) {
	g.startTimes.Store(gameID, time.Now())
}

// ElapsedAndClear returns the elapsed duration since the game started
// and removes the entry. Returns zero and false if the game was not tracked.
func (g *GameTiming) ElapsedAndClear(gameID int64) (time.Duration, bool) {
	value, found := g.startTimes.LoadAndDelete(gameID)
	if !found {
		return 0, false
	}

	startTime, valid := value.(time.Time)
	if !valid {
		return 0, false
	}

	return time.Since(startTime), true
}
