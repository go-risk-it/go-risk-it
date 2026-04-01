package runner

import (
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_MoveSucceeded_RecordsAllMetrics(t *testing.T) {
	t.Parallel()

	c := metrics.NewCollector(0)
	h := &MetricsHandler{collector: c}
	bus := NewTestBus()
	h.Register(bus)

	// Emit MoveDecided first (records phase entry).
	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy},
		UserID: "u0",
		Phase:  "deploy",
	})

	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionDeploy},
		RESTLatency: 50 * time.Millisecond,
		RESTEndTime: time.Now(),
	})

	snap := c.Snapshot()
	assert.Equal(t, int64(1), snap.TotalMoves)
}

func TestMetrics_MoveConflict_RecordsConflict(t *testing.T) {
	t.Parallel()

	c := metrics.NewCollector(0)
	h := &MetricsHandler{collector: c}
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveConflictEvent{Action: &player.Action{Type: player.ActionDeploy}})

	snap := c.Snapshot()
	assert.Equal(t, int64(1), snap.TotalConflicts)
}

func TestMetrics_MoveFailed_RecordsErrorType(t *testing.T) {
	t.Parallel()

	c := metrics.NewCollector(0)
	h := &MetricsHandler{collector: c}
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     assert.AnError,
		ErrType: "stale_state",
	})

	snap := c.Snapshot()
	assert.Equal(t, int64(1), snap.TotalErrors)
	assert.Equal(t, int64(1), snap.ErrorBreakdown["stale_state"])
}

func TestMetrics_GameComplete_Normal(t *testing.T) {
	t.Parallel()

	c := metrics.NewCollector(0)
	h := &MetricsHandler{collector: c}
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameCompleteEvent{Result: GameResult{
		Duration: 5 * time.Second,
		Moves:    10,
		Winner:   "u0",
	}})

	snap := c.Snapshot()
	assert.Equal(t, int64(1), snap.GamesCompleted)
}

func TestMetrics_GameComplete_TimedOut(t *testing.T) {
	t.Parallel()

	c := metrics.NewCollector(0)
	h := &MetricsHandler{collector: c}
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameCompleteEvent{Result: GameResult{
		Duration: 5 * time.Second,
		TimedOut: true,
	}})

	snap := c.Snapshot()
	assert.Equal(t, int64(1), snap.GamesTimedOut)
}

func TestMetrics_GameComplete_Fatal(t *testing.T) {
	t.Parallel()

	c := metrics.NewCollector(0)
	h := &MetricsHandler{collector: c}
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameCompleteEvent{Result: GameResult{
		FatalError: assert.AnError,
	}})

	snap := c.Snapshot()
	assert.Equal(t, int64(1), snap.GamesFatal)
}

func TestMetrics_PhaseEntry_Recorded(t *testing.T) {
	t.Parallel()

	c := metrics.NewCollector(0)
	h := &MetricsHandler{collector: c}
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy},
		UserID: "u0",
		Phase:  "deploy",
	})

	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionAttack},
		UserID: "u0",
		Phase:  "attack",
	})

	snap := c.Snapshot()
	require.Contains(t, snap.PhaseEntries, "deploy")
	require.Contains(t, snap.PhaseEntries, "attack")
	assert.Equal(t, int64(1), snap.PhaseEntries["deploy"])
	assert.Equal(t, int64(1), snap.PhaseEntries["attack"])
}

func TestMetrics_MoveSucceeded_RecordsPhaseMove(t *testing.T) {
	t.Parallel()

	c := metrics.NewCollector(0)
	h := &MetricsHandler{collector: c}
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionAttack},
		RESTLatency: 5 * time.Millisecond,
		RESTEndTime: time.Now(),
	})

	snap := c.Snapshot()
	require.Contains(t, snap.PhaseMoves, "attack")
	assert.Equal(t, int64(1), snap.PhaseMoves["attack"])
	assert.Equal(t, int64(1), snap.TotalMoves)
}

func TestMetrics_GameComplete_ReceivesEventBeforeBusStop(t *testing.T) {
	t.Parallel()

	// Regression test: GameComplete must be seen by MetricsHandler BEFORE
	// the result-capture handler calls Bus.Stop(). This test verifies the
	// handler registration order established in runner.Run().
	c := metrics.NewCollector(0)
	mh := &MetricsHandler{collector: c}
	bus := NewTestBus()

	// Register MetricsHandler first (as wireHandlers does in production).
	mh.Register(bus)

	// Register stop handler AFTER (as runner.Run now does).
	var stopped bool
	bus.On(EventGameComplete, func(b *Bus, _ Event) {
		stopped = true
		b.Stop()
	})

	bus.Emit(GameCompleteEvent{Result: GameResult{
		Duration: 5 * time.Second,
		Moves:    10,
		Winner:   "u0",
	}})

	snap := c.Snapshot()
	assert.True(t, stopped, "stop handler should have fired")
	assert.Equal(t, int64(1), snap.GamesCompleted,
		"MetricsHandler must see GameComplete before Stop()")
}
