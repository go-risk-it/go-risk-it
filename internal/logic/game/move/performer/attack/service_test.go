package attack_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/attack"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	*db.Querier,
	*region.Service,
	*attack.ServiceImpl,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	regionService := region.NewService(t)
	service := attack.NewService(regionService)

	return querier, regionService, service
}

func input() (string, string, ctx.MoveContext) {
	gameID := int64(1)
	userID := "giovanni"
	attackingRegion := "greenland"
	defendingRegion := "iceland"
	userContext := ctx.WithUserID(ctx.WithLog(context.Background(), zap.NewNop().Sugar()), userID)

	gameContext := ctx.WithGameID(userContext, gameID)

	return attackingRegion, defendingRegion, ctx.NewMoveContext(
		userContext,
		gameContext,
	)
}

func TestServiceImpl_AttackShouldFail(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name                 string
		troopsInSource       int64
		troopsInTarget       int64
		attackingTroops      int64
		attackingRegionOwner string
		defendingRegionOwner string
		expectedError        string
	}

	tests := []inputType{
		{
			"When attacking region is not owned by player",
			5,
			5,
			3,
			"gabriele",
			"giovanni",
			"attacking region is not owned by player",
		},
		{
			"When both regions are owned by the same player",
			5,
			5,
			3,
			"giovanni",
			"giovanni",
			"cannot attack your own region",
		},
		{
			"When attacking region has zero troops",
			0,
			5,
			3,
			"giovanni",
			"gabriele",
			"attacking region does not have enough troops",
		},
		{
			"When attacking region does not have enough troops",
			3,
			5,
			3,
			"giovanni",
			"gabriele",
			"attacking region does not have enough troops",
		},
		{
			"When attacking with zero troops",
			3,
			5,
			0,
			"giovanni",
			"gabriele",
			"at least one troop is required to attack",
		},
		{
			"When attacking a region that has zero troops",
			4,
			0,
			3,
			"giovanni",
			"gabriele",
			"defending region does not have enough troops",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, regionService, service := setup(t)
			attackingRegion, defendingRegion, ctx := input()

			game := &sqlc.Game{
				ID:               ctx.GameID(),
				Phase:            sqlc.PhaseDEPLOY,
				Turn:             2,
				DeployableTroops: 5,
			}
			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, attackingRegion).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                1,
					ExternalReference: attackingRegion,
					UserID:            test.attackingRegionOwner,
					Troops:            test.troopsInSource,
				}, nil)
			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, defendingRegion).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                2,
					ExternalReference: defendingRegion,
					UserID:            test.defendingRegionOwner,
					Troops:            test.troopsInTarget,
				}, nil)

			err := service.PerformQ(ctx, querier, game, attack.Move{
				AttackingRegionID: attackingRegion,
				DefendingRegionID: defendingRegion,
				TroopsInSource:    test.troopsInSource,
				TroopsInTarget:    test.troopsInTarget,
				AttackingTroops:   test.attackingTroops,
			})

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
}
