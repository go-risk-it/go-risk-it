package runner //nolint:testpackage // whitebox tests access unexported helpers

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRESTForExecutor tracks which method was called and returns a configurable error.
type fakeRESTForExecutor struct {
	calledMethod string
	calledGameID int64
	err          error
}

func (f *fakeRESTForExecutor) CreateGame(
	_ context.Context,
	_ client.CreateGameRequest,
) (int64, error) {
	return 0, nil
}

func (f *fakeRESTForExecutor) Deploy(_ context.Context, gameID int64, _ client.DeployMove) error {
	f.calledMethod = "deploy"
	f.calledGameID = gameID

	return f.err
}

func (f *fakeRESTForExecutor) Attack(_ context.Context, gameID int64, _ client.AttackMove) error {
	f.calledMethod = "attack"
	f.calledGameID = gameID

	return f.err
}

func (f *fakeRESTForExecutor) Conquer(
	_ context.Context,
	gameID int64,
	_ client.ConquerMove,
) error {
	f.calledMethod = "conquer"
	f.calledGameID = gameID

	return f.err
}

func (f *fakeRESTForExecutor) Reinforce(
	_ context.Context,
	gameID int64,
	_ client.ReinforceMove,
) error {
	f.calledMethod = "reinforce"
	f.calledGameID = gameID

	return f.err
}

func (f *fakeRESTForExecutor) PlayCards(
	_ context.Context,
	gameID int64,
	_ client.CardsMove,
) error {
	f.calledMethod = "cards"
	f.calledGameID = gameID

	return f.err
}

func (f *fakeRESTForExecutor) Advance(_ context.Context, gameID int64, _ string) error {
	f.calledMethod = "advance"
	f.calledGameID = gameID

	return f.err
}

//nolint:unparam // interface conformance / future use
func makeExecutorHandler(rest *fakeRESTForExecutor) (*ExecutorHandler, *GameSession) {
	gameCtx := &GameSession{
		Ctx:    context.Background(),
		GameID: 42,
		Players: []*PlayerInfo{
			{UserID: "u0", Name: "p0", REST: rest},
		},
		UserIndex: map[string]int{"u0": 0},
	}

	return &ExecutorHandler{gameCtx: gameCtx}, gameCtx
}

func TestExecutor_Success_EmitsMoveSucceeded(t *testing.T) {
	t.Parallel()

	rest := &fakeRESTForExecutor{}
	h, _ := makeExecutorHandler(rest)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{
			Type:   player.ActionDeploy,
			Deploy: &player.DeployAction{RegionID: "r1", CurrentTroops: 1, DesiredTroops: 3},
		},
		UserID: "u0",
		Phase:  metrics.PhaseDeploy,
	})

	succeeded := bus.EmittedOfType(EventMoveSucceeded)
	require.Len(t, succeeded, 1)

	//nolint:forcetypeassert // test assertion
	ms := succeeded[0].(MoveSucceededEvent)
	assert.NotNil(t, ms.Action)
	assert.Positive(t, ms.RESTLatency.Nanoseconds())
	assert.False(t, ms.RESTEndTime.IsZero())
}

func TestExecutor_Conflict_EmitsMoveConflict(t *testing.T) {
	t.Parallel()

	rest := &fakeRESTForExecutor{err: &client.ConflictError{Message: "409 conflict"}}
	h, _ := makeExecutorHandler(rest)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy, Deploy: &player.DeployAction{}},
		UserID: "u0",
		Phase:  metrics.PhaseDeploy,
	})

	conflicts := bus.EmittedOfType(EventMoveConflict)
	require.Len(t, conflicts, 1)
	//nolint:forcetypeassert // test assertion
	assert.NotNil(t, conflicts[0].(MoveConflictEvent).Action)
}

func TestExecutor_StaleState_EmitsMoveFailed(t *testing.T) {
	t.Parallel()

	rest := &fakeRESTForExecutor{err: &client.StaleStateError{Message: "stale"}}
	h, _ := makeExecutorHandler(rest)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy, Deploy: &player.DeployAction{}},
		UserID: "u0",
		Phase:  metrics.PhaseDeploy,
	})

	failures := bus.EmittedOfType(EventMoveFailed)
	require.Len(t, failures, 1)
	//nolint:forcetypeassert // test assertion
	assert.Equal(t, metrics.ErrorTypeStaleState, failures[0].(MoveFailedEvent).ErrType)
}

func TestExecutor_Transient_EmitsMoveFailed(t *testing.T) {
	t.Parallel()

	rest := &fakeRESTForExecutor{
		err: &client.TransientError{Cause: errors.New("503"), StatusCode: 503},
	}
	h, _ := makeExecutorHandler(rest)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy, Deploy: &player.DeployAction{}},
		UserID: "u0",
		Phase:  metrics.PhaseDeploy,
	})

	failures := bus.EmittedOfType(EventMoveFailed)
	require.Len(t, failures, 1)
	//nolint:forcetypeassert // test assertion
	assert.Equal(t, metrics.ErrorTypeTransient, failures[0].(MoveFailedEvent).ErrType)
}

func TestExecutor_GenericError_EmitsMoveFailed(t *testing.T) {
	t.Parallel()

	rest := &fakeRESTForExecutor{err: errors.New("boom")}
	h, _ := makeExecutorHandler(rest)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveDecidedEvent{
		Action: &player.Action{Type: player.ActionDeploy, Deploy: &player.DeployAction{}},
		UserID: "u0",
		Phase:  metrics.PhaseDeploy,
	})

	failures := bus.EmittedOfType(EventMoveFailed)
	require.Len(t, failures, 1)
	//nolint:forcetypeassert // test assertion
	assert.Equal(t, metrics.ErrorTypeExecution, failures[0].(MoveFailedEvent).ErrType)
}

func TestExecutor_AllActionTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action *player.Action
		want   string
	}{
		{
			"deploy",
			&player.Action{Type: player.ActionDeploy, Deploy: &player.DeployAction{}},
			"deploy",
		},
		{
			"attack",
			&player.Action{Type: player.ActionAttack, Attack: &player.AttackAction{}},
			"attack",
		},
		{
			"conquer",
			&player.Action{Type: player.ActionConquer, Conquer: &player.ConquerAction{Troops: 1}},
			"conquer",
		},
		{
			"reinforce",
			&player.Action{Type: player.ActionReinforce, Reinforce: &player.ReinforceAction{}},
			"reinforce",
		},
		{
			"cards",
			&player.Action{Type: player.ActionPlayCards, Cards: &player.CardsAction{}},
			"cards",
		},
		{
			"advance",
			&player.Action{
				Type:    player.ActionAdvance,
				Advance: &player.AdvanceAction{CurrentPhase: "deploy"},
			},
			"advance",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rest := &fakeRESTForExecutor{}
			h, _ := makeExecutorHandler(rest)
			bus := NewTestBus()
			h.Register(bus)

			bus.Emit(MoveDecidedEvent{Action: tc.action, UserID: "u0", Phase: metrics.PhaseDeploy})

			assert.Equal(t, tc.want, rest.calledMethod)
			assert.Equal(t, int64(42), rest.calledGameID)
		})
	}
}
