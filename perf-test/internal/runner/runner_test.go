package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/client"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Integration test fakes ---

// fakeStrategyForRunner controls how many moves happen before game-over.
type fakeStrategyForRunner struct {
	movesBeforeGameOver int
	moveCount           atomic.Int32
}

func (f *fakeStrategyForRunner) Name() string { return "fake-runner" }
func (f *fakeStrategyForRunner) DecideMove(
	snap gamestate.ViewSnapshot,
	_ string,
) (*player.Action, error) {
	n := int(f.moveCount.Add(1))
	if n >= f.movesBeforeGameOver {
		// Signal game over by returning advance (the WS view will be updated
		// to show a winner before the next strategy call).
		return &player.Action{
			Type:    player.ActionAdvance,
			Advance: &player.AdvanceAction{CurrentPhase: "deploy"},
		}, nil
	}

	return &player.Action{
		Type:   player.ActionDeploy,
		Deploy: &player.DeployAction{RegionID: "r1", CurrentTroops: 1, DesiredTroops: 3},
	}, nil
}

// fakeWSForRunner simulates WS with controllable game-over.
type fakeWSForRunner struct {
	view       *gamestate.View
	done       chan struct{}
	closeCalls int
	moveCount  *atomic.Int32
	gameOverAt int
}

func newFakeWSForRunner(moveCount *atomic.Int32, gameOverAt int) *fakeWSForRunner {
	v := gamestate.NewView()
	// Initialize with valid state.
	gs := &gamestate.GameState{Turn: 0, Phase: gamestate.Phase{Type: gamestate.Deploy}}
	data, _ := json.Marshal(gs)
	_ = v.Apply(gamestate.WSMessage{Type: "gameState", Payload: data})

	ps := &gamestate.PlayersState{
		Players: []gamestate.Player{
			{UserID: "user-0", Index: 0},
			{UserID: "user-1", Index: 1},
		},
	}
	psData, _ := json.Marshal(ps)
	_ = v.Apply(gamestate.WSMessage{Type: "playerState", Payload: psData})

	return &fakeWSForRunner{
		view:       v,
		done:       make(chan struct{}),
		moveCount:  moveCount,
		gameOverAt: gameOverAt,
	}
}

func (f *fakeWSForRunner) View() *gamestate.View { return f.view }
func (f *fakeWSForRunner) Done() <-chan struct{} { return f.done }
func (f *fakeWSForRunner) Disrupt()              {}

func (f *fakeWSForRunner) Close() error {
	f.closeCalls++

	return nil
}

// simulateUpdate triggers a WS update, optionally setting game-over.
func (f *fakeWSForRunner) simulateUpdate() {
	moves := int(f.moveCount.Load())
	if moves >= f.gameOverAt {
		// Set winner to trigger game over.
		gs := &gamestate.GameState{
			Turn:         int64(moves),
			Phase:        gamestate.Phase{Type: gamestate.Deploy},
			WinnerUserID: "user-0",
		}
		data, _ := json.Marshal(gs)
		_ = f.view.Apply(gamestate.WSMessage{Type: "gameState", Payload: data})
	} else {
		gs := &gamestate.GameState{
			Turn:  int64(moves),
			Phase: gamestate.Phase{Type: gamestate.Deploy},
		}
		data, _ := json.Marshal(gs)
		_ = f.view.Apply(gamestate.WSMessage{Type: "gameState", Payload: data})
	}
}

// fakeRESTForRunner tracks calls and simulates success.
// allWS mirrors production behavior: every REST call updates all WS views.
type fakeRESTForRunner struct {
	deployCalls  int
	advanceCalls int
	allWS        []*fakeWSForRunner
}

func (f *fakeRESTForRunner) CreateGame(
	_ client.CreateGameRequest,
) (int64, error) {
	return 42, nil
}

func (f *fakeRESTForRunner) Deploy(_ int64, _ client.DeployMove) error {
	f.deployCalls++
	// Simulate state update on all WS connections (mirrors server broadcast).
	for _, ws := range f.allWS {
		ws.simulateUpdate()
	}

	return nil
}

func (f *fakeRESTForRunner) Attack(int64, client.AttackMove) error       { return nil }
func (f *fakeRESTForRunner) Conquer(int64, client.ConquerMove) error     { return nil }
func (f *fakeRESTForRunner) Reinforce(int64, client.ReinforceMove) error { return nil }
func (f *fakeRESTForRunner) PlayCards(int64, client.CardsMove) error     { return nil }

func (f *fakeRESTForRunner) Advance(_ int64, _ string) error {
	f.advanceCalls++
	for _, ws := range f.allWS {
		ws.simulateUpdate()
	}

	return nil
}

func TestRunner_HappyPath_CompletesGame(t *testing.T) {
	t.Parallel()

	strategy := &fakeStrategyForRunner{movesBeforeGameOver: 3}
	moveCount := &strategy.moveCount

	ws0 := newFakeWSForRunner(moveCount, 3)
	ws1 := newFakeWSForRunner(moveCount, 3)
	allWS := []*fakeWSForRunner{ws0, ws1}
	rest0 := &fakeRESTForRunner{allWS: allWS}
	rest1 := &fakeRESTForRunner{allWS: allWS}

	cfg := Config{
		BaseURL:   "http://localhost",
		WSURL:     "ws://localhost",
		AnonKey:   "test",
		Strategy:  strategy,
		Timeout:   10 * time.Second,
		Collector: metrics.NewCollector(0),
		Timeouts: Timeouts{
			InitialStateWait:  1 * time.Millisecond,
			UpdateWait:        50 * time.Millisecond,
			PhaseChangeWait:   50 * time.Millisecond,
			PostMoveSettle:    1 * time.Millisecond,
			MaxConsecutiveErr: 20,
		},
	}

	r := newTestRunner(cfg, func(gameCtx *GameContext) {
		gameCtx.GameID = 42
		gameCtx.Players = []*PlayerInfo{
			{
				UserID: "user-0",
				Name:   "bot-0-0",
				REST:   rest0,
				WS:     ws0,
				Auth:   &client.AuthResult{UserID: "user-0"},
			},
			{
				UserID: "user-1",
				Name:   "bot-0-1",
				REST:   rest1,
				WS:     ws1,
				Auth:   &client.AuthResult{UserID: "user-1"},
			},
		}
		gameCtx.UserIndex = map[string]int{"user-0": 0, "user-1": 1}
	})

	ctx := context.Background()
	result := r.Run(ctx, 0, 2)

	assert.Empty(t, result.FatalError)
	assert.Equal(t, "user-0", result.Winner)
	assert.GreaterOrEqual(t, result.Moves, 0)
}

func TestRunner_SetupFailure_ReturnsFatal(t *testing.T) {
	t.Parallel()

	cfg := Config{
		BaseURL:   "http://localhost",
		WSURL:     "ws://localhost",
		AnonKey:   "test",
		Strategy:  &fakeStrategy{action: &player.Action{Type: player.ActionDeploy}},
		Timeout:   5 * time.Second,
		Collector: metrics.NewCollector(0),
		Timeouts:  DefaultTimeouts(),
	}

	r := New(cfg)
	// Override auth to fail.
	r.protocolFactory = func(gameCtx *GameContext) *ProtocolHandler {
		return &ProtocolHandler{
			baseURL:  cfg.BaseURL,
			wsURL:    cfg.WSURL,
			anonKey:  cfg.AnonKey,
			timeouts: cfg.Timeouts,
			gameCtx:  gameCtx,
			newAuth: func(_, _ string) AuthClient {
				return &fakeAuth{failAt: 0, err: fmt.Errorf("signup failed")}
			},
			newREST: func(_, _ string, _ *metrics.Collector) RESTClient { return nil },
			newWS: func(_ string, _ int64, _ string, _ *metrics.Collector) (WSClient, error) {
				return nil, nil
			},
		}
	}

	ctx := context.Background()
	result := r.Run(ctx, 0, 4)

	require.Error(t, result.FatalError)
	assert.Contains(t, result.FatalError.Error(), "signup")
}

func TestRunner_Timeout_ReturnsTimedOut(t *testing.T) {
	t.Parallel()

	// Strategy that never ends the game.
	neverEndStrategy := &fakeStrategy{
		action: &player.Action{
			Type:   player.ActionDeploy,
			Deploy: &player.DeployAction{RegionID: "r1", CurrentTroops: 1, DesiredTroops: 3},
		},
	}

	ws := newFakeWSForRunner(&atomic.Int32{}, 999999)
	rest := &fakeRESTForRunner{allWS: []*fakeWSForRunner{ws}}

	cfg := Config{
		BaseURL:   "http://localhost",
		WSURL:     "ws://localhost",
		AnonKey:   "test",
		Strategy:  neverEndStrategy,
		Timeout:   50 * time.Millisecond, // Very short timeout.
		Collector: metrics.NewCollector(0),
		Timeouts: Timeouts{
			InitialStateWait:  1 * time.Millisecond,
			UpdateWait:        10 * time.Millisecond,
			PhaseChangeWait:   10 * time.Millisecond,
			PostMoveSettle:    1 * time.Millisecond,
			MaxConsecutiveErr: 20,
		},
	}

	r := newTestRunner(cfg, func(gameCtx *GameContext) {
		gameCtx.GameID = 42
		gameCtx.Players = []*PlayerInfo{
			{
				UserID: "user-0",
				Name:   "bot-0-0",
				REST:   rest,
				WS:     ws,
				Auth:   &client.AuthResult{UserID: "user-0"},
			},
		}
		gameCtx.UserIndex = map[string]int{"user-0": 0}
	})

	ctx := context.Background()
	result := r.Run(ctx, 0, 1)

	// Should terminate due to timeout.
	assert.True(t, result.TimedOut || result.FatalError != nil,
		"expected timeout or fatal, got: %+v", result)
}

func TestRunner_WSConnectionsClosed(t *testing.T) {
	t.Parallel()

	strategy := &fakeStrategyForRunner{movesBeforeGameOver: 1}
	moveCount := &strategy.moveCount

	ws0 := newFakeWSForRunner(moveCount, 1)
	ws1 := newFakeWSForRunner(moveCount, 1)
	allWS := []*fakeWSForRunner{ws0, ws1}
	rest := &fakeRESTForRunner{allWS: allWS}

	cfg := Config{
		BaseURL:   "http://localhost",
		WSURL:     "ws://localhost",
		AnonKey:   "test",
		Strategy:  strategy,
		Timeout:   5 * time.Second,
		Collector: metrics.NewCollector(0),
		Timeouts: Timeouts{
			InitialStateWait:  1 * time.Millisecond,
			UpdateWait:        50 * time.Millisecond,
			PhaseChangeWait:   50 * time.Millisecond,
			PostMoveSettle:    1 * time.Millisecond,
			MaxConsecutiveErr: 20,
		},
	}

	r := newTestRunner(cfg, func(gameCtx *GameContext) {
		gameCtx.GameID = 42
		gameCtx.Players = []*PlayerInfo{
			{
				UserID: "user-0",
				Name:   "bot-0-0",
				REST:   rest,
				WS:     ws0,
				Auth:   &client.AuthResult{UserID: "user-0"},
			},
			{
				UserID: "user-1",
				Name:   "bot-0-1",
				REST:   rest,
				WS:     ws1,
				Auth:   &client.AuthResult{UserID: "user-1"},
			},
		}
		gameCtx.UserIndex = map[string]int{"user-0": 0, "user-1": 1}
	})

	ctx := context.Background()
	_ = r.Run(ctx, 0, 2)

	assert.Equal(t, 1, ws0.closeCalls)
	assert.Equal(t, 1, ws1.closeCalls)
}

func TestRunner_RunFuncSignature(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Collector: metrics.NewCollector(0),
		Timeouts:  DefaultTimeouts(),
	}
	r := New(cfg)

	// Verify ToRunFunc returns the right type.
	var _ orchestrator.RunFunc = r.ToRunFunc()
}
