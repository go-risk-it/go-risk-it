package controller_test

import (
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game"
	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	ctx2 "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	conquerMock "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/conquer"
	deployMock "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/deploy"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func phaseGameContext(t *testing.T) ctx2.GameContext {
	t.Helper()

	return ctx2.WithGameID(
		ctx2.WithUserID(
			ctx2.WithSpan(t.Context(), noop.Span{}),
			"testuser",
		),
		42,
	)
}

func newPhaseController(t *testing.T) (
	*controller.PhaseController,
	*deployMock.Service,
	*conquerMock.Service,
) {
	t.Helper()

	deploySvc := deployMock.NewService(t)
	conquerSvc := conquerMock.NewService(t)
	ctrl := controller.NewPhaseController(conquerSvc, deploySvc)

	return ctrl, deploySvc, conquerSvc
}

func TestPhaseController_GetAttackPhaseState(t *testing.T) {
	t.Parallel()

	ctrl, _, _ := newPhaseController(t)
	ctx := phaseGameContext(t)
	gameState := &state.Game{ID: 42, Turn: 7, WinnerUserID: ""}

	result, err := ctrl.GetAttackPhaseState(ctx, gameState)

	require.NoError(t, err)
	require.Equal(t, messaging.GameState[messaging.EmptyState]{
		ID:   42,
		Turn: 7,
		Phase: messaging.Phase[messaging.EmptyState]{
			Type:  game.Attack,
			State: messaging.EmptyState{},
		},
		WinnerUserID: "",
	}, result)
}

func TestPhaseController_GetReinforcePhaseState(t *testing.T) {
	t.Parallel()

	ctrl, _, _ := newPhaseController(t)
	ctx := phaseGameContext(t)
	gameState := &state.Game{ID: 42, Turn: 7, WinnerUserID: "winner-user"}

	result, err := ctrl.GetReinforcePhaseState(ctx, gameState)

	require.NoError(t, err)
	require.Equal(t, messaging.GameState[messaging.EmptyState]{
		ID:   42,
		Turn: 7,
		Phase: messaging.Phase[messaging.EmptyState]{
			Type:  game.Reinforce,
			State: messaging.EmptyState{},
		},
		WinnerUserID: "winner-user",
	}, result)
}

func TestPhaseController_GetCardsPhaseState(t *testing.T) {
	t.Parallel()

	ctrl, _, _ := newPhaseController(t)
	ctx := phaseGameContext(t)
	gameState := &state.Game{ID: 42, Turn: 7, WinnerUserID: ""}

	result, err := ctrl.GetCardsPhaseState(ctx, gameState)

	require.NoError(t, err)
	require.Equal(t, messaging.GameState[messaging.EmptyState]{
		ID:   42,
		Turn: 7,
		Phase: messaging.Phase[messaging.EmptyState]{
			Type:  game.Cards,
			State: messaging.EmptyState{},
		},
		WinnerUserID: "",
	}, result)
}

func TestPhaseController_GetDeployPhaseState(t *testing.T) {
	t.Parallel()

	ctrl, deploySvc, _ := newPhaseController(t)
	ctx := phaseGameContext(t)
	gameState := &state.Game{ID: 42, Turn: 3, WinnerUserID: ""}

	deploySvc.EXPECT().GetDeployableTroops(ctx).Return(int64(5), nil)

	result, err := ctrl.GetDeployPhaseState(ctx, gameState)

	require.NoError(t, err)
	require.Equal(t, messaging.GameState[messaging.DeployPhaseState]{
		ID:   42,
		Turn: 3,
		Phase: messaging.Phase[messaging.DeployPhaseState]{
			Type: game.Deploy,
			State: messaging.DeployPhaseState{
				DeployableTroops: 5,
			},
		},
		WinnerUserID: "",
	}, result)
}

func TestPhaseController_GetDeployPhaseState_Error(t *testing.T) {
	t.Parallel()

	ctrl, deploySvc, _ := newPhaseController(t)
	ctx := phaseGameContext(t)
	gameState := &state.Game{ID: 42, Turn: 3, WinnerUserID: ""}

	deploySvc.EXPECT().GetDeployableTroops(ctx).Return(int64(0), errors.New("db down"))

	_, err := ctrl.GetDeployPhaseState(ctx, gameState)

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get deployable troops")
	require.ErrorContains(t, err, "db down")
}

func TestPhaseController_GetConquerPhaseState(t *testing.T) {
	t.Parallel()

	ctrl, _, conquerSvc := newPhaseController(t)
	ctx := phaseGameContext(t)
	gameState := &state.Game{ID: 42, Turn: 5, WinnerUserID: ""}

	conquerSvc.EXPECT().GetPhaseState(ctx).Return(sqlc.GetConquerPhaseStateRow{
		SourceRegion:  "greenland",
		TargetRegion:  "iceland",
		MinimumTroops: 3,
	}, nil)

	result, err := ctrl.GetConquerPhaseState(ctx, gameState)

	require.NoError(t, err)
	require.Equal(t, messaging.GameState[messaging.ConquerPhaseState]{
		ID:   42,
		Turn: 5,
		Phase: messaging.Phase[messaging.ConquerPhaseState]{
			Type: game.Conquer,
			State: messaging.ConquerPhaseState{
				AttackingRegionID: "greenland",
				DefendingRegionID: "iceland",
				MinTroopsToMove:   3,
			},
		},
		WinnerUserID: "",
	}, result)
}

func TestPhaseController_GetConquerPhaseState_Error(t *testing.T) {
	t.Parallel()

	ctrl, _, conquerSvc := newPhaseController(t)
	ctx := phaseGameContext(t)
	gameState := &state.Game{ID: 42, Turn: 5, WinnerUserID: ""}

	conquerSvc.EXPECT().
		GetPhaseState(ctx).
		Return(sqlc.GetConquerPhaseStateRow{}, errors.New("db down"))

	_, err := ctrl.GetConquerPhaseState(ctx, gameState)

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to get conquer phase state")
	require.ErrorContains(t, err, "db down")
}
