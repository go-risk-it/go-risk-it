package deploy_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/player"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	gamestate "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/state"
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
	phaseService := phase.NewService(t)
	regionService := region.NewService(t)
	service := deploy.NewService(querier, gameService, phaseService, playerService, regionService)

	return querier, playerService, gameService, regionService, service
}

func input() (string, int64, int64, ctx.GameContext) {
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
		), gameContext
}

func TestServiceImpl_DeployShouldFailWhenPlayerDoesntHaveEnoughDeployableTroops(t *testing.T) {
	t.Parallel()

	querier, _, _, _, service := setup(t)
	regionReference, currentTroops, desiredTroops, ctx := input()

	game := &state.Game{
		ID:    ctx.GameID(),
		Phase: sqlc.PhaseTypeDEPLOY,
		Turn:  2,
	}

	querier.EXPECT().GetDeployableTroops(ctx, game.ID).Return(int64(0), nil)

	_, err := service.PerformQ(ctx, querier, deploy.Move{
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
			"region is not owned by player",
		},
		{
			"When amount of troops declared is wrong",
			"Giovanni",
			10,
			"region has different number of troops than declared",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, _, _, regionService, service := setup(t)
			regionReference, _, desiredTroops, ctx := input()

			currentTroops := test.declaredTroops

			game := &state.Game{
				ID:    ctx.GameID(),
				Phase: sqlc.PhaseTypeDEPLOY,
				Turn:  2,
			}

			querier.EXPECT().GetDeployableTroops(ctx, game.ID).Return(int64(5), nil)

			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, regionReference).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                1,
					ExternalReference: "greenland",
					UserID:            test.regionOwner,
					Troops:            0,
				}, nil)

			_, err := service.PerformQ(ctx, querier, deploy.Move{
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
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, playerService, gameService, regionService, service := setup(
				t,
			)
			regionReference, currentTroops, desiredTroops, ctx := input()
			troops := desiredTroops - currentTroops

			game := &state.Game{
				ID:    ctx.GameID(),
				Phase: sqlc.PhaseTypeDEPLOY,
				Turn:  2,
			}

			querier.EXPECT().GetDeployableTroops(ctx, game.ID).Return(test.deployableTroops, nil)

			region := &sqlc.GetRegionsByGameRow{
				ID:                1,
				ExternalReference: "greenland",
				UserID:            "Giovanni",
				Troops:            currentTroops,
			}

			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, regionReference).
				Return(region, nil)
			regionService.
				EXPECT().
				UpdateTroopsInRegionQ(ctx, querier, region, troops).
				Return(nil)
			querier.
				EXPECT().
				DecreaseDeployableTroops(ctx, sqlc.DecreaseDeployableTroopsParams{
					ID:               ctx.GameID(),
					DeployableTroops: desiredTroops - currentTroops,
				}).
				Return(nil)

			_, err := service.PerformQ(ctx, querier, deploy.Move{
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
