package attack_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/board"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/move/attack/dice"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	*db.Querier,
	*board.Service,
	*dice.Service,
	*region.Service,
	*attack.ServiceImpl,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	boardService := board.NewService(t)
	diceService := dice.NewService(t)
	phaseService := phase.NewService(t)
	regionService := region.NewService(t)
	service := attack.NewService(boardService, diceService, phaseService, regionService)

	return querier, boardService, diceService, regionService, service
}

func input() ctx.MoveContext {
	gameID := int64(1)
	userID := "giovanni"

	userContext := ctx.WithUserID(
		ctx.WithLog(context.Background(), zap.NewExample().Sugar()),
		userID,
	)

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
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, boardService, _, regionService, service := setup(t)
			ctx := input()

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

			_, err := service.PerformQ(ctx, querier, attack.Move{
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

func TestServiceImpl_AttackShouldUpdateRegionTroops(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name                       string
		attackingTroops            int64
		defendingTroops            int64
		troopsInDefendingRegion    int64
		attackDices                []int
		defenseDices               []int
		expectedAttackerCasualties int64
		expectedDefenderCasualties int64
	}

	tests := []inputType{
		{
			"When one attack dice is strictly worse",
			1,
			2,
			2,
			[]int{1},
			[]int{2, 3},
			1,
			0,
		},
		{
			"When one attack dice is equal or worse",
			1,
			2,
			2,
			[]int{2},
			[]int{2, 3},
			1,
			0,
		},
		{
			"When one attacker dice is better than a defender, but worse than the corresponding",
			2,
			2,
			2,
			[]int{3, 5},
			[]int{4, 5},
			2,
			0,
		},
		{
			"When both have losses",
			2,
			2,
			2,
			[]int{3, 5},
			[]int{2, 5},
			1,
			1,
		},
		{
			"When attacker wins all",
			2,
			2,
			2,
			[]int{3, 5},
			[]int{2, 4},
			0,
			2,
		},
		{
			"When in tie, defender wins",
			2,
			2,
			2,
			[]int{3, 5},
			[]int{3, 5},
			2,
			0,
		},
		{
			"When attacking with less than 3 troops and region has at least 3, should outnumber",
			2,
			3,
			3,
			[]int{3, 5},
			[]int{3, 4, 5},
			2,
			0,
		},
		{
			"When defending region has more than 3 troops",
			3,
			3,
			10,
			[]int{3, 5, 6},
			[]int{3, 4, 5},
			1,
			2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, boardService, diceService, regionService, service := setup(t)
			ctx := input()

			troopsInAttackingRegion := int64(4)
			attackingRegion := &sqlc.GetRegionsByGameRow{
				ID:                1,
				ExternalReference: "greenland",
				UserID:            "giovanni",
				Troops:            troopsInAttackingRegion,
			}
			defendingRegion := &sqlc.GetRegionsByGameRow{
				ID:                2,
				ExternalReference: "iceland",
				UserID:            "gabriele",
				Troops:            test.troopsInDefendingRegion,
			}

			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, attackingRegion.ExternalReference).
				Return(attackingRegion, nil)
			regionService.
				EXPECT().
				GetRegionQ(ctx, querier, defendingRegion.ExternalReference).
				Return(defendingRegion, nil)
			boardService.
				EXPECT().
				AreNeighbours(ctx, attackingRegion.ExternalReference, defendingRegion.ExternalReference).
				Return(true, nil)
			diceService.
				EXPECT().
				RollAttackingDices(len(test.attackDices)).
				Return(test.attackDices).
				Once()
			diceService.
				EXPECT().
				RollDefendingDices(len(test.defenseDices)).
				Return(test.defenseDices).
				Once()
			regionService.
				EXPECT().
				UpdateTroopsInRegionQ(ctx, querier, attackingRegion, -test.expectedAttackerCasualties).
				Return(nil)
			regionService.
				EXPECT().
				UpdateTroopsInRegionQ(ctx, querier, defendingRegion, -test.expectedDefenderCasualties).
				Return(nil)

			_, err := service.PerformQ(ctx, querier, attack.Move{
				AttackingRegionID: attackingRegion.ExternalReference,
				DefendingRegionID: defendingRegion.ExternalReference,
				TroopsInSource:    troopsInAttackingRegion,
				TroopsInTarget:    test.troopsInDefendingRegion,
				AttackingTroops:   test.attackingTroops,
			})

			require.NoError(t, err)
		})
	}
}

func TestServiceImpl_HasConqueredQ(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name     string
		regions  []sqlc.GetRegionsByGameRow
		expected bool
	}

	tests := []inputType{
		{
			"When nothing was conquered",
			[]sqlc.GetRegionsByGameRow{
				{
					ID:                1,
					ExternalReference: "greenland",
					UserID:            "giovanni",
					Troops:            4,
				},
				{
					ID:                2,
					ExternalReference: "iceland",
					UserID:            "gabriele",
					Troops:            5,
				},
			},
			false,
		},
		{
			"When a region is conquered, but it's owned by the attacker",
			[]sqlc.GetRegionsByGameRow{
				{
					ID:                1,
					ExternalReference: "greenland",
					UserID:            "giovanni",
					Troops:            4,
				},
				{
					ID:                2,
					ExternalReference: "iceland",
					UserID:            "giovanni",
					Troops:            0,
				},
			},
			false,
		},
		{
			"When a region is conquered, and it's not owned by the attacker",
			[]sqlc.GetRegionsByGameRow{
				{
					ID:                1,
					ExternalReference: "greenland",
					UserID:            "giovanni",
					Troops:            4,
				},
				{
					ID:                2,
					ExternalReference: "iceland",
					UserID:            "gabriele",
					Troops:            0,
				},
			},
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, _, _, regionService, service := setup(t)
			ctx := input()

			regionService.
				EXPECT().
				GetRegionsQ(ctx, querier).
				Return(test.regions, nil)

			result, err := service.HasConqueredQ(ctx, querier)

			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestServiceImpl_CanContinueAttackingQ(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name     string
		regions  []sqlc.GetRegionsByGameRow
		expected bool
	}

	tests := []inputType{
		{
			"When player still has troops",
			[]sqlc.GetRegionsByGameRow{
				{
					ID:                1,
					ExternalReference: "greenland",
					UserID:            "giovanni",
					Troops:            4,
				},
				{
					ID:                2,
					ExternalReference: "iceland",
					UserID:            "gabriele",
					Troops:            5,
				},
			},
			true,
		},
		{
			"When player doesn't have troops anymore",
			[]sqlc.GetRegionsByGameRow{
				{
					ID:                1,
					ExternalReference: "greenland",
					UserID:            "giovanni",
					Troops:            1,
				},
				{
					ID:                2,
					ExternalReference: "iceland",
					UserID:            "gabriele",
					Troops:            2,
				},
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, _, _, regionService, service := setup(t)
			ctx := input()

			regionService.
				EXPECT().
				GetRegionsQ(ctx, querier).
				Return(test.regions, nil)

			result, err := service.CanContinueAttackingQ(ctx, querier)

			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		})
	}
}
