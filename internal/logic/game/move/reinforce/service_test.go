package reinforce_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	realcards "github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/board"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	mockstate "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/state"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (
	*db.Querier,
	*board.Service,
	*cards.Service,
	*mockstate.Service,
	*phase.Service,
	*region.Service,
	reinforce.Service,
) {
	t.Helper()

	querier := db.NewQuerier(t)
	boardService := board.NewService(t)
	cardsService := cards.NewService(t)
	gameService := mockstate.NewService(t)
	phaseService := phase.NewService(t)
	regionService := region.NewService(t)
	service, _ := reinforce.NewService(
		boardService,
		cardsService,
		gameService,
		phaseService,
		regionService,
	)

	return querier, boardService, cardsService, gameService, phaseService, regionService, service
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
		name                string
		sourceRegionID      string
		targetRegionID      string
		declaredTroopsInSrc int64
		declaredTroopsInTgt int64
		troopsInSource      int64
		troopsInTarget      int64
		movingTroops        int64
		sourceOwner         string
		targetOwner         string
		canReach            bool
		expectedError       string
	}

	tests := []inputType{
		{
			"When source region is not owned by player",
			"greenland",
			"iceland",
			5,
			3,
			5,
			3,
			2,
			"gabriele",
			"giovanni",
			true,
			"validation failed: region ownership check failed: source region is not owned by player",
		},
		{
			"When target region is not owned by player",
			"greenland",
			"iceland",
			5,
			3,
			5,
			3,
			2,
			"giovanni",
			"gabriele",
			true,
			"validation failed: region ownership check failed: target region is not owned by player",
		},
		{
			"When moving zero troops",
			"greenland",
			"iceland",
			5,
			3,
			5,
			3,
			0,
			"giovanni",
			"giovanni",
			true,
			"validation failed: troops check failed: at least one troop is required to reinforce",
		},
		{
			"When source region does not have enough troops",
			"greenland",
			"iceland",
			5,
			3,
			5,
			3,
			5,
			"giovanni",
			"giovanni",
			true,
			"validation failed: troops check failed: source region does not have enough troops",
		},
		{
			"When source region doesn't have the declared number of troops",
			"greenland",
			"iceland",
			4,
			3,
			5,
			3,
			2,
			"giovanni",
			"giovanni",
			true,
			"validation failed: troops check failed: declared values are invalid: source region doesn't have the declared number of troops",
		},
		{
			"When target region doesn't have the declared number of troops",
			"greenland",
			"iceland",
			5,
			4,
			5,
			3,
			2,
			"giovanni",
			"giovanni",
			true,
			"validation failed: troops check failed: declared values are invalid: target region doesn't have the declared number of troops",
		},
		{
			"When regions are not reachable",
			"greenland",
			"siam",
			5,
			3,
			5,
			3,
			2,
			"giovanni",
			"giovanni",
			false,
			"validation failed: player cannot reach target region",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, boardService, _, _, _, regionService, service := setup(t)
			ctx := input()

			regionService.
				EXPECT().
				GetRegion(ctx, querier, test.sourceRegionID).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                1,
					ExternalReference: test.sourceRegionID,
					UserID:            test.sourceOwner,
					Troops:            test.troopsInSource,
				}, nil)
			regionService.
				EXPECT().
				GetRegion(ctx, querier, test.targetRegionID).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                2,
					ExternalReference: test.targetRegionID,
					UserID:            test.targetOwner,
					Troops:            test.troopsInTarget,
				}, nil)

			if !test.canReach {
				boardService.
					EXPECT().
					CanPlayerReach(ctx, querier, test.sourceRegionID, test.targetRegionID).
					Return(false, nil)
			}

			_, err := service.Perform(ctx, querier, reinforce.Move{
				SourceRegionID: test.sourceRegionID,
				TargetRegionID: test.targetRegionID,
				TroopsInSource: test.declaredTroopsInSrc,
				TroopsInTarget: test.declaredTroopsInTgt,
				MovingTroops:   test.movingTroops,
			})

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestServiceImpl_Perform_ShouldMoveTroops(t *testing.T) {
	t.Parallel()

	querier, boardService, _, _, _, regionService, service := setup(t)
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
		UserID:            "giovanni",
		Troops:            3,
	}

	regionService.
		EXPECT().
		GetRegion(ctx, querier, "greenland").
		Return(sourceRegion, nil)
	regionService.
		EXPECT().
		GetRegion(ctx, querier, "iceland").
		Return(targetRegion, nil)
	boardService.
		EXPECT().
		CanPlayerReach(ctx, querier, "greenland", "iceland").
		Return(true, nil)
	regionService.
		EXPECT().
		UpdateTroopsInRegion(ctx, querier, sourceRegion, int64(-3)).
		Return(nil)
	regionService.
		EXPECT().
		UpdateTroopsInRegion(ctx, querier, targetRegion, int64(3)).
		Return(nil)

	_, err := service.Perform(ctx, querier, reinforce.Move{
		SourceRegionID: "greenland",
		TargetRegionID: "iceland",
		TroopsInSource: 5,
		TroopsInTarget: 3,
		MovingTroops:   3,
	})

	require.NoError(t, err)
}

func TestServiceImpl_Walk(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name                string
		hasValidCombination bool
		expectedNextPhase   sqlc.GamePhaseType
	}

	tests := []inputType{
		{
			"When next player has valid combination, should return CARDS",
			true,
			sqlc.GamePhaseTypeCARDS,
		},
		{
			"When next player has no valid combination, should return DEPLOY",
			false,
			sqlc.GamePhaseTypeDEPLOY,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, _, cardsService, _, _, _, service := setup(t)
			ctx := input()

			cardsService.
				EXPECT().
				NextPlayerHasValidCombination(ctx, querier).
				Return(test.hasValidCombination, nil)

			nextPhase, err := service.Walk(ctx, querier, false)

			require.NoError(t, err)
			require.Equal(t, test.expectedNextPhase, nextPhase)
		})
	}
}

func TestServiceImpl_Advance_ToCards_WithConquerInTurn(t *testing.T) {
	t.Parallel()

	querier, _, cardsService, gameService, phaseService, _, service := setup(t)
	ctx := input()

	gameService.
		EXPECT().
		GetGameStateWithQuerier(ctx, querier).
		Return(&state.Game{
			ID:   1,
			Turn: 3,
		}, nil)
	querier.
		EXPECT().
		HasConqueredInTurn(ctx, sqlc.HasConqueredInTurnParams{
			ID:   ctx.GameID(),
			Turn: 3,
		}).
		Return(true, nil)
	cardsService.
		EXPECT().
		Draw(ctx, querier).
		Return(nil)
	phaseService.
		EXPECT().
		InsertPhase(ctx, querier, sqlc.GamePhaseTypeCARDS).
		Return(&sqlc.GamePhase{}, nil)

	err := service.Advance(ctx, querier, sqlc.GamePhaseTypeCARDS, struct{}{})

	require.NoError(t, err)
}

func TestServiceImpl_Advance_ToCards_WithoutConquerInTurn(t *testing.T) {
	t.Parallel()

	querier, _, _, gameService, phaseService, _, service := setup(t)
	ctx := input()

	gameService.
		EXPECT().
		GetGameStateWithQuerier(ctx, querier).
		Return(&state.Game{
			ID:   1,
			Turn: 3,
		}, nil)
	querier.
		EXPECT().
		HasConqueredInTurn(ctx, sqlc.HasConqueredInTurnParams{
			ID:   ctx.GameID(),
			Turn: 3,
		}).
		Return(false, nil)
	phaseService.
		EXPECT().
		InsertPhase(ctx, querier, sqlc.GamePhaseTypeCARDS).
		Return(&sqlc.GamePhase{}, nil)

	err := service.Advance(ctx, querier, sqlc.GamePhaseTypeCARDS, struct{}{})

	require.NoError(t, err)
}

func TestServiceImpl_Advance_ToDeploy_WithConquerInTurn(t *testing.T) {
	t.Parallel()

	querier, _, cardsService, gameService, _, _, service := setup(t)
	ctx := input()

	gameService.
		EXPECT().
		GetGameStateWithQuerier(ctx, querier).
		Return(&state.Game{
			ID:   1,
			Turn: 5,
		}, nil)
	querier.
		EXPECT().
		HasConqueredInTurn(ctx, sqlc.HasConqueredInTurnParams{
			ID:   ctx.GameID(),
			Turn: 5,
		}).
		Return(true, nil)
	cardsService.
		EXPECT().
		Draw(ctx, querier).
		Return(nil)
	cardsService.
		EXPECT().
		Advance(ctx, querier, sqlc.GamePhaseTypeDEPLOY, (*realcards.MoveResult)(nil)).
		Return(nil)

	err := service.Advance(ctx, querier, sqlc.GamePhaseTypeDEPLOY, struct{}{})

	require.NoError(t, err)
}

func TestServiceImpl_Advance_ToDeploy_WithoutConquerInTurn(t *testing.T) {
	t.Parallel()

	querier, _, cardsService, gameService, _, _, service := setup(t)
	ctx := input()

	gameService.
		EXPECT().
		GetGameStateWithQuerier(ctx, querier).
		Return(&state.Game{
			ID:   1,
			Turn: 5,
		}, nil)
	querier.
		EXPECT().
		HasConqueredInTurn(ctx, sqlc.HasConqueredInTurnParams{
			ID:   ctx.GameID(),
			Turn: 5,
		}).
		Return(false, nil)
	cardsService.
		EXPECT().
		Advance(ctx, querier, sqlc.GamePhaseTypeDEPLOY, (*realcards.MoveResult)(nil)).
		Return(nil)

	err := service.Advance(ctx, querier, sqlc.GamePhaseTypeDEPLOY, struct{}{})

	require.NoError(t, err)
}

func TestServiceImpl_Advance_InvalidTransition(t *testing.T) {
	t.Parallel()

	querier, _, _, _, _, _, service := setup(t)
	ctx := input()

	err := service.Advance(ctx, querier, sqlc.GamePhaseTypeATTACK, struct{}{})

	require.Error(t, err)
	require.ErrorContains(t, err, "invalid phase transition")
}

func TestServiceImpl_PhaseType(t *testing.T) {
	t.Parallel()

	_, _, _, _, _, _, service := setup(t)

	require.Equal(t, sqlc.GamePhaseTypeREINFORCE, service.PhaseType())
}
