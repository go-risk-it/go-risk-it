package reinforce_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/reinforce"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/region"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (
	*db.Querier,
	*board.Service,
	*region.Service,
	reinforce.Service,
) {
	t.Helper()

	querier := db.NewQuerier(t)
	boardService := board.NewService(t)
	cardsService := cards.NewService(t)
	phaseService := phase.NewService(t)
	regionService := region.NewService(t)
	service, _ := reinforce.NewService(
		boardService,
		cardsService,
		phaseService,
		regionService,
	)

	return querier, boardService, regionService, service
}

func input() ctx.GameContext {
	gameID := int64(1)
	userID := "giovanni"

	userContext := kernelctx.WithUserID(
		kernelctx.WithSpan(context.Background(), noop.Span{}),
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

			querier, boardService, regionService, service := setup(t)
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

			_, _, err := service.Perform(ctx, querier, reinforce.Move{
				SourceRegionID: test.sourceRegionID,
				TargetRegionID: test.targetRegionID,
				TroopsInSource: test.declaredTroopsInSrc,
				TroopsInTarget: test.declaredTroopsInTgt,
				MovingTroops:   test.movingTroops,
			}, nil)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestServiceImpl_Perform_ShouldMoveTroops(t *testing.T) {
	t.Parallel()

	querier, boardService, regionService, service := setup(t)
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

	_, effect, err := service.Perform(ctx, querier, reinforce.Move{
		SourceRegionID: "greenland",
		TargetRegionID: "iceland",
		TroopsInSource: 5,
		TroopsInTarget: 3,
		MovingTroops:   3,
	}, nil)

	require.NoError(t, err)

	// Verify MoveEffect
	require.Len(t, effect.RegionUpdates, 2)
	require.Equal(t, moveservice.RegionUpdate{
		RegionID:  "greenland",
		NewOwner:  "giovanni",
		NewTroops: 2,
	}, effect.RegionUpdates[0])
	require.Equal(t, moveservice.RegionUpdate{
		RegionID:  "iceland",
		NewOwner:  "giovanni",
		NewTroops: 6,
	}, effect.RegionUpdates[1])
	require.Equal(t, snapshot.EmptyPhaseState{}, effect.UpdatedPhase)
}
