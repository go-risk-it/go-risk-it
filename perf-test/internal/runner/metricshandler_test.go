package runner

import (
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_MoveSucceeded_RecordsAllMetrics(t *testing.T) {
	t.Parallel()

	c := metrics.NewCollector(0)
	h := &MetricsHandler{collector: c}
	bus := NewTestBus()
	h.Register(bus)

	// Emit MoveDecided first (sets moveStartTime + phase).
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

func TestMetrics_PhaseEntry_OnChange(t *testing.T) {
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
	// Both "deploy" and "attack" should have phase entries.
	require.Contains(t, snap.PhaseEntries, "deploy")
	require.Contains(t, snap.PhaseEntries, "attack")
	assert.Equal(t, int64(1), snap.PhaseEntries["deploy"])
	assert.Equal(t, int64(1), snap.PhaseEntries["attack"])
}

func TestMetrics_StateReceived_RecordsWSDelivery(t *testing.T) {
	t.Parallel()

	c := metrics.NewCollector(0)
	h := &MetricsHandler{collector: c}
	bus := NewTestBus()
	h.Register(bus)

	// Set lastRESTEndTime by emitting a MoveSucceeded.
	restEnd := time.Now()
	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionDeploy},
		RESTLatency: time.Millisecond,
		RESTEndTime: restEnd,
	})

	// Now emit StateReceived after the REST end.
	bus.Emit(StateReceivedEvent{
		Snapshot:  gamestate.ViewSnapshot{},
		Timestamp: restEnd.Add(10 * time.Millisecond),
	})

	// WS delivery should have been recorded.
	snap := c.Snapshot()
	assert.Greater(t, snap.WSDelivery.Count, int64(0))
}
