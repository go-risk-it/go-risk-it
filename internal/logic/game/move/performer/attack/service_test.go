package attack_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/attack"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/board"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	*db.Querier,
	*board.Service,
	*region.Service,
	*attack.ServiceImpl,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	boardService := board.NewService(t)
	regionService := region.NewService(t)
	service := attack.NewService(boardService, regionService)

	return querier, boardService, regionService, service
}

func input() ctx.MoveContext {
	gameID := int64(1)
	userID := "giovanni"
	userContext := ctx.WithUserID(ctx.WithLog(context.Background(), zap.NewNop().Sugar()), userID)

	gameContext := ctx.WithGameID(userContext, gameID)

	return ctx.NewMoveContext(
		userContext,
		gameContext,
	)
}

func TestServiceImpl_AttackShouldFail(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name                   string
		attackingRegion        string
		defendingRegion        string
		declaredTroopsInSource int64
		declaredTroopsInTarget int64
		troopsInSource         int64
		troopsInTarget         int64
		attackingTroops        int64
		attackingRegionOwner   string
		defendingRegionOwner   string
		regionsAreNeighboring  bool
		expectedError          string
	}

	tests := []inputType{
		{
			"When attacking region is not owned by player",
			"greenland",
			"iceland",
			5,
			5,
			5,
			5,
			3,
			"gabriele",
			"giovanni",
			true,
			"validation failed: region ownership check failed: attacking region is not owned by player",
		},
		{
			"When both regions are owned by the same player",
			"greenland",
			"iceland",
			5,
			5,
			5,
			5,
			3,
			"giovanni",
			"giovanni",
			true,
			"validation failed: region ownership check failed: cannot attack your own region",
		},
		{
			"When attacking region has zero troops",
			"greenland",
			"iceland",
			0,
			5,
			0,
			5,
			3,
			"giovanni",
			"gabriele",
			true,
			"validation failed: troops check failed: attacking region does not have enough troops",
		},
		{
			"When attacking region does not have enough troops",
			"greenland",
			"iceland",
			3,
			5,
			3,
			5,
			3,
			"giovanni",
			"gabriele",
			true,
			"validation failed: troops check failed: attacking region does not have enough troops",
		},
		{
			"When attacking with zero troops",
			"greenland",
			"iceland",
			3,
			5,
			3,
			5,
			0,
			"giovanni",
			"gabriele",
			true,
			"validation failed: troops check failed: at least one troop is required to attack",
		},
		{
			"When attacking a region that has zero troops",
			"greenland",
			"iceland",
			4,
			0,
			4,
			0,
			3,
			"giovanni",
			"gabriele",
			true,
			"validation failed: troops check failed: defending region does not have enough troops",
		},
		{
			"When attacking region doesn't have the declared number of troops",
			"greenland",
			"iceland",
			4,
			3,
			5,
			3,
			3,
			"giovanni",
			"gabriele",
			true,
			"validation failed: troops check failed: declared values are invalid: attacking region doesn't have the declared number of troops",
		},
		{
			"When defending region doesn't have the declared number of troops",
			"greenland",
			"iceland",
			4,
			3,
			4,
			4,
			3,
			"giovanni",
			"gabriele",
			true,
			"validation failed: troops check failed: declared values are invalid: defending region doesn't have the declared number of troops",
		},
		{
			"When attacking and defending regions are not neighbours",
			"greenland",
			"siam",
			4,
			3,
			4,
			3,
			3,
			"giovanni",
			"gabriele",
			false,
			"validation failed: attacking region cannot reach defending region",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, boardService, regionService, service := setup(t)
			ctx := input()

			game := &sqlc.Game{
				ID:               ctx.GameID(),
				Phase:            sqlc.PhaseDEPLOY,
				Turn:             2,
				DeployableTroops: 5,
			}
			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, test.attackingRegion).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                1,
					ExternalReference: test.attackingRegion,
					UserID:            test.attackingRegionOwner,
					Troops:            test.troopsInSource,
				}, nil)
			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, test.defendingRegion).
				Return(&sqlc.GetRegionsByGameRow{
					ID:                2,
					ExternalReference: test.defendingRegion,
					UserID:            test.defendingRegionOwner,
					Troops:            test.troopsInTarget,
				}, nil)
			if !test.regionsAreNeighboring {
				boardService.
					EXPECT().
					AreNeighbours(ctx, test.attackingRegion, test.defendingRegion).
					Return(false, nil)
			}

			err := service.PerformQ(ctx, querier, game, attack.Move{
				AttackingRegionID: test.attackingRegion,
				DefendingRegionID: test.defendingRegion,
				TroopsInSource:    test.declaredTroopsInSource,
				TroopsInTarget:    test.declaredTroopsInTarget,
				AttackingTroops:   test.attackingTroops,
			})

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
}
