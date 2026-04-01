package runner

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeStateWatcherHandler(
	ctx context.Context,
) (*StateWatcherHandler, *GameSession) {
	snap := mkSnap(0, gamestate.Deploy, "")
	ws0 := newFakeWSWithState(snap)

	gameCtx := &GameSession{
		Ctx:       ctx,
		GameIndex: 1,
		Players: []*PlayerInfo{
			{UserID: "u0", Name: "p0", WS: ws0},
		},
		UserIndex: map[string]int{"u0": 0},
	}

	h := &StateWatcherHandler{
		gameCtx: gameCtx,
		timeouts: Timeouts{
			UpdateWait:      100 * time.Millisecond,
			PhaseChangeWait: 100 * time.Millisecond,
			PostMoveSettle:  10 * time.Millisecond,
		},
	}

	return h, gameCtx
}

func TestStateWatcher_MoveSucceeded_EmitsState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h, gameCtx := makeStateWatcherHandler(ctx)
	bus := NewTestBus()
	h.Register(bus)

	// Signal an update on the WS view so waitForAnyUpdate returns quickly.
	go func() {
		time.Sleep(10 * time.Millisecond)
		data, _ := json.Marshal(
			&gamestate.GameState{Turn: 1, Phase: gamestate.Phase{Type: gamestate.Deploy}},
		)
		_ = gameCtx.Players[0].WS.View().
			Apply(gamestate.WSMessage{Type: "gameState", Payload: data})
	}()

	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionDeploy},
		RESTLatency: time.Millisecond,
		RESTEndTime: time.Now(),
	})

	stateEvents := bus.EmittedOfType(EventStateReceived)
	require.Len(t, stateEvents, 1)
}

func TestStateWatcher_TurnSkipped_EmitsState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h, gameCtx := makeStateWatcherHandler(ctx)
	bus := NewTestBus()
	h.Register(bus)

	go func() {
		time.Sleep(10 * time.Millisecond)
		data, _ := json.Marshal(
			&gamestate.GameState{Turn: 1, Phase: gamestate.Phase{Type: gamestate.Deploy}},
		)
		_ = gameCtx.Players[0].WS.View().
			Apply(gamestate.WSMessage{Type: "gameState", Payload: data})
	}()

	bus.Emit(TurnSkippedEvent{})

	stateEvents := bus.EmittedOfType(EventStateReceived)
	require.Len(t, stateEvents, 1)
}

func TestStateWatcher_MoveConflict_WaitsForPhaseChange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h, gameCtx := makeStateWatcherHandler(ctx)
	bus := NewTestBus()
	h.Register(bus)

	// Simulate a phase change after a short delay.
	go func() {
		time.Sleep(10 * time.Millisecond)
		data, _ := json.Marshal(
			&gamestate.GameState{Turn: 0, Phase: gamestate.Phase{Type: gamestate.Attack}},
		)
		_ = gameCtx.Players[0].WS.View().
			Apply(gamestate.WSMessage{Type: "gameState", Payload: data})
	}()

	bus.Emit(MoveConflictEvent{
		Action: &player.Action{Type: player.ActionDeploy},
	})

	stateEvents := bus.EmittedOfType(EventStateReceived)
	require.Len(t, stateEvents, 1)
}

func TestStateWatcher_Timeout_StillEmitsCurrentState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h, _ := makeStateWatcherHandler(ctx)
	// Very short timeout so we hit it.
	h.timeouts.UpdateWait = 10 * time.Millisecond
	bus := NewTestBus()
	h.Register(bus)

	// Don't signal any update — will timeout.
	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionDeploy},
		RESTLatency: time.Millisecond,
		RESTEndTime: time.Now(),
	})

	stateEvents := bus.EmittedOfType(EventStateReceived)
	require.Len(t, stateEvents, 1, "should emit current state even on timeout")
}

func TestStateWatcher_ContextCancelled_NoEmit(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	h, _ := makeStateWatcherHandler(ctx)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionDeploy},
		RESTLatency: time.Millisecond,
		RESTEndTime: time.Now(),
	})

	stateEvents := bus.EmittedOfType(EventStateReceived)
	assert.Empty(t, stateEvents)
}
