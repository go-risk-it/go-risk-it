package runner //nolint:testpackage // whitebox tests access unexported helpers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeStateWatcherHandler(
	ctx context.Context,
) (*StateWatcherHandler, *GameSession) {
	snap := mkSnap(0, snapshot.PhaseDeploy, "")
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
		},
	}

	return h, gameCtx
}

func signalWSUpdatePlayerView(t *testing.T, p *PlayerInfo, pv *snapshot.PlayerView) {
	t.Helper()

	// Ensure Mission is set for JSON round-trip (PlayerMission.UnmarshalJSON
	// errors on empty type).
	if pv.Mission.Type == "" {
		pv.Mission = snapshot.PlayerMission{
			Type:   snapshot.MissionTwentyFourTerritories,
			Detail: snapshot.TwentyFourTerritoriesMission{},
		}
	}

	data, err := json.Marshal(pv)
	assert.NoError(t, err) //nolint:testifylint // Called from test goroutines - safe with assert

	_ = p.WS.View().Apply(gamestate.WSMessage{Type: "playerView", Payload: data})
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
		signalWSUpdatePlayerView(t, gameCtx.Players[0], &snapshot.PlayerView{
			Game: snapshot.GameMeta{Turn: 1},
			Phase: snapshot.Phase{
				Type:  snapshot.PhaseDeploy,
				State: snapshot.DeployPhaseState{DeployableTroops: 3},
			},
		})
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
		signalWSUpdatePlayerView(t, gameCtx.Players[0], &snapshot.PlayerView{
			Game: snapshot.GameMeta{Turn: 1},
			Phase: snapshot.Phase{
				Type:  snapshot.PhaseDeploy,
				State: snapshot.DeployPhaseState{DeployableTroops: 3},
			},
		})
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
		signalWSUpdatePlayerView(t, gameCtx.Players[0], &snapshot.PlayerView{
			Game: snapshot.GameMeta{Turn: 0},
			Phase: snapshot.Phase{
				Type:  snapshot.PhaseAttack,
				State: snapshot.EmptyPhaseState{},
			},
		})
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
