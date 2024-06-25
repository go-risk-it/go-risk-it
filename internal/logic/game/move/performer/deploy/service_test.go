package deploy_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/deploy"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/gamestate"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/player"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	*db.Querier,
	*player.Service,
	*gamestate.Service,
	*region.Service,
	*deploy.ServiceImpl,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	playerService := player.NewService(t)
	gameService := gamestate.NewService(t)
	regionService := region.NewService(t)
	service := deploy.NewService(
		querier,
		zap.NewNop().Sugar(),
		gameService,
		playerService,
		regionService,
	)

	return querier, playerService, gameService, regionService, service
}

func input() (string, int64, int64, ctx.MoveContext) {
	gameID := int64(1)
	userID := "Giovanni"
	regionReference := "greenland"
	currentTroops := 0
	desiredTroops := 5
	userContext := ctx.WithUserID(ctx.WithLog(context.Background(), zap.NewNop().Sugar()), userID)

	gameContext := ctx.WithGameID(userContext, gameID)

	return regionReference, int64(
			currentTroops,
		), int64(
			desiredTroops,
		), ctx.NewMoveContext(
			userContext,
			gameContext,
		)
}

func TestServiceImpl_DeployShouldFailWhenPlayerDoesntHaveEnoughDeployableTroops(t *testing.T) {
	t.Parallel()

	querier, _, _, _, service := setup(t)
	regionReference, currentTroops, desiredTroops, ctx := input()

	game := &sqlc.Game{
		ID:    ctx.GameID(),
		Phase: sqlc.PhaseDEPLOY,
		Turn:  2,
	}
	err := service.PerformQ(ctx, querier, game, deploy.Move{
		RegionID:      regionReference,
		CurrentTroops: currentTroops,
		DesiredTroops: desiredTroops,
	})

	require.Error(t, err)
	require.EqualError(t, err, "not enough deployable troops")
}

func TestServiceImpl_DeployShouldFail(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name           string
		regionOwner    string
		declaredTroops int64
		expectedError  string
	}

	tests := []inputType{
		{
			"When region is not owned by player",
			"Gabriele",
			0,
			"failed to get region: region is not owned by player",
		},
		{
			"When amount of troops declared is wrong",
			"Giovanni",
			10,
			"region has different number of troops than declared",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, _, _, regionService, service := setup(t)
			regionReference, _, desiredTroops, ctx := input()

			currentTroops := test.declaredTroops

			game := &sqlc.Game{
				ID:               ctx.GameID(),
				Phase:            sqlc.PhaseDEPLOY,
				Turn:             2,
				DeployableTroops: 5,
			}
			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, ctx.GameID(), regionReference).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                1,
					ExternalReference: "greenland",
					UserID:            test.regionOwner,
					Troops:            0,
				}, nil)

			err := service.PerformQ(ctx, querier, game, deploy.Move{
				RegionID:      regionReference,
				CurrentTroops: currentTroops,
				DesiredTroops: desiredTroops,
			})

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestServiceImpl_DeployShouldSucceed(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name             string
		deployableTroops int64
	}

	tests := []inputType{
		{
			"Should succeed without advancing phase",
			15,
		},
		{
			"Should succeed and advance phase",
			5,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, playerService, gameService, regionService, service := setup(
				t,
			)
			regionReference, currentTroops, desiredTroops, ctx := input()
			troops := desiredTroops - currentTroops

			game := &sqlc.Game{
				ID:               ctx.GameID(),
				Phase:            sqlc.PhaseDEPLOY,
				Turn:             2,
				DeployableTroops: test.deployableTroops,
			}
			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, ctx.GameID(), regionReference).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                1,
					ExternalReference: "greenland",
					UserID:            "Giovanni",
					Troops:            0,
				}, nil)
			gameService.
				EXPECT().
				DecreaseDeployableTroopsQ(ctx, querier, game, troops).
				Return(nil)
			regionService.
				EXPECT().
				IncreaseTroopsInRegion(ctx, querier, int64(1), troops).
				Return(nil)

			err := service.PerformQ(ctx, querier, game, deploy.Move{
				RegionID:      regionReference,
				CurrentTroops: currentTroops,
				DesiredTroops: desiredTroops,
			})

			require.NoError(t, err)
			gameService.AssertExpectations(t)
			playerService.AssertExpectations(t)
			regionService.AssertExpectations(t)
		})
	}
}
