package runner

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeStrategy implements player.Strategy for testing.
type fakeStrategy struct {
	action *player.Action
	err    error
}

func (f *fakeStrategy) Name() string { return "fake" }
func (f *fakeStrategy) DecideMove(_ gamestate.ViewSnapshot, _ string) (*player.Action, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.action, nil
}

func makeStrategyHandler(
	ctx context.Context,
	strategy player.Strategy,
) (*StrategyHandler, *GameSession) {
	// Default state for the WS views — deploy phase, turn 0.
	defaultSnap := mkSnap(0, gamestate.Deploy, "")

	gameCtx := &GameSession{
		GameIndex: 1,
		Players: []*PlayerInfo{
			{UserID: "u0", Name: "p0", WS: newFakeWSWithState(defaultSnap)},
			{UserID: "u1", Name: "p1", WS: newFakeWSWithState(defaultSnap)},
			{UserID: "u2", Name: "p2", WS: newFakeWSWithState(defaultSnap)},
		},
		UserIndex: map[string]int{"u0": 0, "u1": 1, "u2": 2},
	}

	h := &StrategyHandler{
		strategy: strategy,
		gameCtx:  gameCtx,
		ctx:      ctx,
	}

	return h, gameCtx
}

func mkSnap(turn int64, phase gamestate.PhaseType, winner string) gamestate.ViewSnapshot {
	gs := &gamestate.GameState{
		Turn:         turn,
		Phase:        gamestate.Phase{Type: phase},
		WinnerUserID: winner,
	}
	ps := &gamestate.PlayersState{
		Players: []gamestate.Player{
			{UserID: "u0", Index: 0},
			{UserID: "u1", Index: 1},
			{UserID: "u2", Index: 2},
		},
	}

	return gamestate.ViewSnapshot{GameState: gs, PlayersState: ps}
}

func TestStrategy_GameOver_EmitsGameComplete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h, _ := makeStrategyHandler(ctx, &fakeStrategy{})
	bus := NewTestBus()
	h.Register(bus)

	snap := mkSnap(0, gamestate.Deploy, "u1")
	bus.Emit(StateReceivedEvent{Snapshot: snap, Timestamp: time.Now()})

	completes := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completes, 1)
	assert.Equal(t, "u1", completes[0].(GameCompleteEvent).Result.Winner)
}

func TestStrategy_NilGameState_EmitsTurnSkipped(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h, _ := makeStrategyHandler(ctx, &fakeStrategy{})
	bus := NewTestBus()
	h.Register(bus)

	snap := gamestate.ViewSnapshot{GameState: nil, PlayersState: nil}
	bus.Emit(StateReceivedEvent{Snapshot: snap, Timestamp: time.Now()})

	assert.Len(t, bus.EmittedOfType(EventTurnSkipped), 1)
	assert.Empty(t, bus.EmittedOfType(EventGameComplete))
}

func TestStrategy_NilPlayersState_EmitsTurnSkipped(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h, _ := makeStrategyHandler(ctx, &fakeStrategy{})
	bus := NewTestBus()
	h.Register(bus)

	snap := gamestate.ViewSnapshot{
		GameState: &gamestate.GameState{
			Turn:  0,
			Phase: gamestate.Phase{Type: gamestate.Deploy},
		},
		PlayersState: nil,
	}
	bus.Emit(StateReceivedEvent{Snapshot: snap, Timestamp: time.Now()})

	assert.Len(t, bus.EmittedOfType(EventTurnSkipped), 1)
}

func TestStrategy_NotMyTurn_EmitsTurnSkipped(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h, _ := makeStrategyHandler(ctx, &fakeStrategy{})
	// Override UserIndex so no player matches.
	h.gameCtx.UserIndex = map[string]int{"unknown": 0}
	bus := NewTestBus()
	h.Register(bus)

	snap := mkSnap(0, gamestate.Deploy, "")
	bus.Emit(StateReceivedEvent{Snapshot: snap, Timestamp: time.Now()})

	assert.Len(t, bus.EmittedOfType(EventTurnSkipped), 1)
}

func TestStrategy_DecideMove_EmitsMoveDecided(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	action := &player.Action{
		Type:   player.ActionDeploy,
		Deploy: &player.DeployAction{RegionID: "r1", CurrentTroops: 1, DesiredTroops: 3},
	}
	h, _ := makeStrategyHandler(ctx, &fakeStrategy{action: action})
	bus := NewTestBus()
	h.Register(bus)

	snap := mkSnap(0, gamestate.Deploy, "")
	bus.Emit(StateReceivedEvent{Snapshot: snap, Timestamp: time.Now()})

	moves := bus.EmittedOfType(EventMoveDecided)
	require.Len(t, moves, 1)

	md := moves[0].(MoveDecidedEvent)
	assert.Equal(t, action, md.Action)
	assert.Equal(t, "u0", md.UserID)
	assert.Equal(t, metrics.Phase("deploy"), md.Phase)
}

func TestStrategy_DecideError_EmitsMoveFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h, _ := makeStrategyHandler(ctx, &fakeStrategy{err: fmt.Errorf("bad strategy")})
	bus := NewTestBus()
	h.Register(bus)

	snap := mkSnap(0, gamestate.Deploy, "")
	bus.Emit(StateReceivedEvent{Snapshot: snap, Timestamp: time.Now()})

	failures := bus.EmittedOfType(EventMoveFailed)
	require.Len(t, failures, 1)

	mf := failures[0].(MoveFailedEvent)
	assert.False(t, mf.Fatal)
	assert.Equal(t, metrics.ErrorTypeStrategy, mf.ErrType)
}

func TestStrategy_ContextCancelled_EmitsGameComplete(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	h, _ := makeStrategyHandler(ctx, &fakeStrategy{})
	bus := NewTestBus()
	h.Register(bus)

	snap := mkSnap(0, gamestate.Deploy, "")
	bus.Emit(StateReceivedEvent{Snapshot: snap, Timestamp: time.Now()})

	completes := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completes, 1)
	assert.False(t, completes[0].(GameCompleteEvent).Result.TimedOut)
}

func TestStrategy_ContextDeadlineExceeded_EmitsTimedOut(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	defer cancel()

	h, _ := makeStrategyHandler(ctx, &fakeStrategy{})
	bus := NewTestBus()
	h.Register(bus)

	snap := mkSnap(0, gamestate.Deploy, "")
	bus.Emit(StateReceivedEvent{Snapshot: snap, Timestamp: time.Now()})

	completes := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completes, 1)
	assert.True(t, completes[0].(GameCompleteEvent).Result.TimedOut)
}

func TestStrategy_FindActivePlayer(t *testing.T) {
	t.Parallel()

	mkPlayers := func(ids ...string) []*PlayerInfo {
		ps := make([]*PlayerInfo, len(ids))
		for i, id := range ids {
			ps[i] = &PlayerInfo{UserID: id, Name: "p" + id}
		}

		return ps
	}

	mkIndex := func(ps []*PlayerInfo) map[string]int {
		idx := make(map[string]int)
		for i, p := range ps {
			idx[p.UserID] = i
		}

		return idx
	}

	type entry struct {
		userID string
		index  int64
	}

	mkPlayersState := func(entries ...entry) *gamestate.PlayersState {
		ps := &gamestate.PlayersState{Players: make([]gamestate.Player, len(entries))}
		for i, e := range entries {
			ps.Players[i] = gamestate.Player{UserID: e.userID, Index: e.index}
		}

		return ps
	}

	players3 := mkPlayers("u0", "u1", "u2")
	index3 := mkIndex(players3)
	ps3 := mkPlayersState(
		entry{"u0", 0},
		entry{"u1", 1},
		entry{"u2", 2},
	)

	tests := []struct {
		name      string
		snap      gamestate.ViewSnapshot
		userIndex map[string]int
		wantIdx   int
		wantNil   bool
	}{
		{
			name: "turn 0 returns player 0",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 0},
				PlayersState: ps3,
			},
			userIndex: index3,
			wantIdx:   0,
		},
		{
			name: "turn 1 returns player 1",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 1},
				PlayersState: ps3,
			},
			userIndex: index3,
			wantIdx:   1,
		},
		{
			name: "turn 3 wraps to player 0",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 3},
				PlayersState: ps3,
			},
			userIndex: index3,
			wantIdx:   0,
		},
		{
			name: "turn 7 wraps to player 1",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 7},
				PlayersState: ps3,
			},
			userIndex: index3,
			wantIdx:   1,
		},
		{
			name:      "nil gameState returns nil",
			snap:      gamestate.ViewSnapshot{GameState: nil, PlayersState: ps3},
			userIndex: index3,
			wantNil:   true,
		},
		{
			name: "nil playersState returns nil",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 0},
				PlayersState: nil,
			},
			userIndex: index3,
			wantNil:   true,
		},
		{
			name: "empty players returns nil",
			snap: gamestate.ViewSnapshot{
				GameState:    &gamestate.GameState{Turn: 0},
				PlayersState: &gamestate.PlayersState{Players: []gamestate.Player{}},
			},
			userIndex: index3,
			wantNil:   true,
		},
		{
			name: "unknown user returns nil",
			snap: gamestate.ViewSnapshot{
				GameState: &gamestate.GameState{Turn: 0},
				PlayersState: &gamestate.PlayersState{
					Players: []gamestate.Player{{UserID: "unknown", Index: 0}},
				},
			},
			userIndex: index3,
			wantNil:   true,
		},
		{
			name:      "both nil returns nil",
			snap:      gamestate.ViewSnapshot{GameState: nil, PlayersState: nil},
			userIndex: index3,
			wantNil:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			idx, userID := findActivePlayer(tc.snap, tc.userIndex)
			if tc.wantNil {
				assert.Equal(t, -1, idx)
				assert.Empty(t, userID)
			} else {
				assert.Equal(t, tc.wantIdx, idx)
				assert.Equal(t, players3[tc.wantIdx].UserID, userID)
			}
		})
	}
}

func TestActionTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action player.ActionType
		want   string
	}{
		{player.ActionDeploy, "deploy"},
		{player.ActionAttack, "attack"},
		{player.ActionConquer, "conquer"},
		{player.ActionReinforce, "reinforce"},
		{player.ActionPlayCards, "cards"},
		{player.ActionAdvance, "advance"},
		{player.ActionType(999), "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, actionTypeName(tc.action))
		})
	}
}
