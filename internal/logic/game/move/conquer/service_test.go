package conquer_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/card"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (
	*db.Querier,
	*attack.Service,
	*card.Service,
	*mission.Service,
	*region.Service,
	conquer.Service,
) {
	t.Helper()

	querier := db.NewQuerier(t)
	attackService := attack.NewService(t)
	cardService := card.NewService(t)
	missionService := mission.NewService(t)
	phaseService := phase.NewService(t)
	regionService := region.NewService(t)
	service, _ := conquer.NewService(
		querier,
		attackService,
		cardService,
		missionService,
		phaseService,
		regionService,
	)

	return querier, attackService, cardService, missionService, regionService, service
}

func input() ctx.GameContext {
	gameID := int64(1)
	userID := "giovanni"

	userContext := ctx.WithUserID(
		ctx.WithSpan(context.Background(), noop.Span{}),
		userID,
	)

	return ctx.WithGameID(userContext, gameID)
}

func TestServiceImpl_Perform_ShouldFailValidation(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name          string
		minimumTroops int64
		moveTroops    int64
		sourceTroops  int64
		expectedError string
	}

	tests := []inputType{
		{
			"When moving fewer troops than minimum",
			3,
			2,
			5,
			"must move at least 3 troops",
		},
		{
			"When source region does not have enough troops",
			2,
			5,
			5,
			"source region does not have enough troops",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, _, _, _, regionService, service := setup(t)
			ctx := input()

			querier.
				EXPECT().
				GetConquerPhaseState(ctx, ctx.GameID()).
				Return(sqlc.GetConquerPhaseStateRow{
					SourceRegion:  "greenland",
					TargetRegion:  "iceland",
					MinimumTroops: test.minimumTroops,
				}, nil)

			if test.minimumTroops <= test.moveTroops {
				regionService.
					EXPECT().
					GetRegion(ctx, querier, "greenland").
					Return(&sqlc.GetRegionsByGameRow{
						ID:                1,
						ExternalReference: "greenland",
						UserID:            "giovanni",
						Troops:            test.sourceTroops,
					}, nil)

				regionService.
					EXPECT().
					GetRegion(ctx, querier, "iceland").
					Return(&sqlc.GetRegionsByGameRow{
						ID:                2,
						ExternalReference: "iceland",
						UserID:            "gabriele",
						Troops:            0,
					}, nil)
			}

			_, err := service.Perform(ctx, querier, conquer.Move{
				Troops: test.moveTroops,
			})

			require.Error(t, err)
			require.ErrorContains(t, err, test.expectedError)
		})
	}
}

func TestServiceImpl_Perform_ShouldTransferTroopsAndOwnership(t *testing.T) {
	t.Parallel()

	querier, _, _, _, regionService, service := setup(t)
	ctx := input()

	sourceRegion := &sqlc.GetRegionsByGameRow{
		ID:                1,
		ExternalReference: "greenland",
		UserID:            "giovanni",
		Troops:            5,
	}
	targetRegion := &sqlc.GetRegionsByGameRow{
		ID:                2,
		ExternalReference: "iceland",
		UserID:            "gabriele",
		Troops:            0,
	}
	defeatedPlayerID := int64(42)

	querier.
		EXPECT().
		GetConquerPhaseState(ctx, ctx.GameID()).
		Return(sqlc.GetConquerPhaseStateRow{
			SourceRegion:  "greenland",
			TargetRegion:  "iceland",
			MinimumTroops: 2,
		}, nil)

	regionService.
		EXPECT().
		GetRegion(ctx, querier, "greenland").
		Return(sourceRegion, nil)
	regionService.
		EXPECT().
		GetRegion(ctx, querier, "iceland").
		Return(targetRegion, nil)
	regionService.
		EXPECT().
		UpdateTroopsInRegion(ctx, querier, sourceRegion, int64(-3)).
		Return(nil)
	regionService.
		EXPECT().
		UpdateTroopsInRegion(ctx, querier, targetRegion, int64(3)).
		Return(nil)
	regionService.
		EXPECT().
		UpdateRegionOwner(ctx, querier, targetRegion).
		Return(defeatedPlayerID, nil)
	regionService.
		EXPECT().
		GetRegionsControlledByPlayer(ctx, querier, defeatedPlayerID).
		Return([]sqlc.GameRegion{
			{ID: 10},
		}, nil)

	_, err := service.Perform(ctx, querier, conquer.Move{
		Troops: 3,
	})

	require.NoError(t, err)
}

func TestServiceImpl_Perform_ShouldHandlePlayerElimination(t *testing.T) {
	t.Parallel()

	querier, _, cardService, missionService, regionService, service := setup(t)
	ctx := input()

	sourceRegion := &sqlc.GetRegionsByGameRow{
		ID:                1,
		ExternalReference: "greenland",
		UserID:            "giovanni",
		Troops:            5,
	}
	targetRegion := &sqlc.GetRegionsByGameRow{
		ID:                2,
		ExternalReference: "iceland",
		UserID:            "gabriele",
		Troops:            0,
	}
	defeatedPlayerID := int64(42)

	querier.
		EXPECT().
		GetConquerPhaseState(ctx, ctx.GameID()).
		Return(sqlc.GetConquerPhaseStateRow{
			SourceRegion:  "greenland",
			TargetRegion:  "iceland",
			MinimumTroops: 2,
		}, nil)

	regionService.
		EXPECT().
		GetRegion(ctx, querier, "greenland").
		Return(sourceRegion, nil)
	regionService.
		EXPECT().
		GetRegion(ctx, querier, "iceland").
		Return(targetRegion, nil)
	regionService.
		EXPECT().
		UpdateTroopsInRegion(ctx, querier, sourceRegion, int64(-3)).
		Return(nil)
	regionService.
		EXPECT().
		UpdateTroopsInRegion(ctx, querier, targetRegion, int64(3)).
		Return(nil)
	regionService.
		EXPECT().
		UpdateRegionOwner(ctx, querier, targetRegion).
		Return(defeatedPlayerID, nil)
	regionService.
		EXPECT().
		GetRegionsControlledByPlayer(ctx, querier, defeatedPlayerID).
		Return([]sqlc.GameRegion{}, nil)
	cardService.
		EXPECT().
		TransferCardsOwnership(ctx, querier, defeatedPlayerID).
		Return(nil)
	missionService.
		EXPECT().
		ReassignMissions(ctx, querier, defeatedPlayerID).
		Return(nil)

	_, err := service.Perform(ctx, querier, conquer.Move{
		Troops: 3,
	})

	require.NoError(t, err)
}

func TestServiceImpl_Walk(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name                  string
		canContinueAttacking  bool
		expectedNextPhaseType sqlc.GamePhaseType
	}

	tests := []inputType{
		{
			"When player can continue attacking, should return ATTACK",
			true,
			sqlc.GamePhaseTypeATTACK,
		},
		{
			"When player cannot continue attacking, should return REINFORCE",
			false,
			sqlc.GamePhaseTypeREINFORCE,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, attackService, _, _, _, service := setup(t)
			ctx := input()

			attackService.
				EXPECT().
				CanContinueAttacking(ctx, querier).
				Return(test.canContinueAttacking, nil)

			nextPhase, err := service.Walk(ctx, querier, false)

			require.NoError(t, err)
			require.Equal(t, test.expectedNextPhaseType, nextPhase)
		})
	}
}

func TestServiceImpl_Advance_ShouldCreatePhaseForValidTransition(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	phaseService := phase.NewService(t)
	service, _ := conquer.NewService(
		querier,
		attack.NewService(t),
		card.NewService(t),
		mission.NewService(t),
		phaseService,
		region.NewService(t),
	)
	ctx := input()

	phaseService.
		EXPECT().
		InsertPhase(ctx, querier, sqlc.GamePhaseTypeATTACK).
		Return(&sqlc.GamePhase{}, nil)

	err := service.Advance(ctx, querier, sqlc.GamePhaseTypeATTACK, struct{}{})

	require.NoError(t, err)
}

func TestServiceImpl_Advance_ShouldFailForInvalidTransition(t *testing.T) {
	t.Parallel()

	querier, _, _, _, _, service := setup(t)
	ctx := input()

	err := service.Advance(ctx, querier, sqlc.GamePhaseTypeDEPLOY, struct{}{})

	require.Error(t, err)
	require.ErrorContains(t, err, "invalid phase transition")
}

func TestServiceImpl_GetPhaseStateWithQuerier(t *testing.T) {
	t.Parallel()

	querier, _, _, _, _, service := setup(t)
	ctx := input()

	expected := sqlc.GetConquerPhaseStateRow{
		SourceRegion:  "greenland",
		TargetRegion:  "iceland",
		MinimumTroops: 3,
	}

	querier.
		EXPECT().
		GetConquerPhaseState(ctx, ctx.GameID()).
		Return(expected, nil)

	result, err := service.GetPhaseStateWithQuerier(ctx, querier)

	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestServiceImpl_PhaseType(t *testing.T) {
	t.Parallel()

	_, _, _, _, _, service := setup(t)

	require.Equal(t, sqlc.GamePhaseTypeCONQUER, service.PhaseType())
}
