package attack_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/attack"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/move/attack/dice"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (
	*board.Service,
	*dice.Service,
	attack.Service,
) {
	t.Helper()
	boardService := board.NewService(t)
	diceService := dice.NewService(t)
	service, _ := attack.NewService(
		boardService,
		diceService,
	)

	return boardService, diceService, service
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
			"validation failed: troops check failed: declared values are invalid: source region doesn't have the declared number of troops",
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
			"validation failed: troops check failed: declared values are invalid: target region doesn't have the declared number of troops",
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

			boardService, _, service := setup(t)
			ctx := input()

			prev := &snapshot.CachedGameState{
				PublicSnapshot: &snapshot.GameSnapshot{
					Regions: []snapshot.RegionState{
						{
							InternalID: 1,
							ID:         test.attackingRegion,
							OwnerID:    test.attackingRegionOwner,
							Troops:     test.troopsInSource,
						},
						{
							InternalID: 2,
							ID:         test.defendingRegion,
							OwnerID:    test.defendingRegionOwner,
							Troops:     test.troopsInTarget,
						},
					},
				},
			}

			if !test.regionsAreNeighboring {
				boardService.
					EXPECT().
					AreNeighbours(ctx, test.attackingRegion, test.defendingRegion).
					Return(false, nil)
			}

			_, _, err := service.Perform(ctx, attack.Move{
				AttackingRegionID: test.attackingRegion,
				DefendingRegionID: test.defendingRegion,
				TroopsInSource:    test.declaredTroopsInSource,
				TroopsInTarget:    test.declaredTroopsInTarget,
				AttackingTroops:   test.attackingTroops,
			}, prev)

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

			boardService, diceService, service := setup(t)
			ctx := input()

			troopsInAttackingRegion := int64(4)
			attackingRegionID := "greenland"
			defendingRegionID := "iceland"

			prev := &snapshot.CachedGameState{
				PublicSnapshot: &snapshot.GameSnapshot{
					Regions: []snapshot.RegionState{
						{
							InternalID: 1,
							ID:         attackingRegionID,
							OwnerID:    "giovanni",
							Troops:     troopsInAttackingRegion,
						},
						{
							InternalID: 2,
							ID:         defendingRegionID,
							OwnerID:    "gabriele",
							Troops:     test.troopsInDefendingRegion,
						},
					},
				},
			}

			boardService.
				EXPECT().
				AreNeighbours(ctx, attackingRegionID, defendingRegionID).
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

			result, effect, err := service.Perform(ctx, attack.Move{
				AttackingRegionID: attackingRegionID,
				DefendingRegionID: defendingRegionID,
				TroopsInSource:    troopsInAttackingRegion,
				TroopsInTarget:    test.troopsInDefendingRegion,
				AttackingTroops:   test.attackingTroops,
			}, prev)

			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify MoveEffect
			require.Len(t, effect.RegionUpdates, 2)
			require.Equal(t, moveservice.RegionUpdate{
				RegionID:  "greenland",
				NewOwner:  "giovanni",
				NewTroops: troopsInAttackingRegion - test.expectedAttackerCasualties,
			}, effect.RegionUpdates[0])
			require.Equal(t, moveservice.RegionUpdate{
				RegionID:  "iceland",
				NewOwner:  "gabriele",
				NewTroops: test.troopsInDefendingRegion - test.expectedDefenderCasualties,
			}, effect.RegionUpdates[1])
			require.Equal(t, snapshot.EmptyPhaseState{}, effect.UpdatedPhase)
		})
	}
}

func TestServiceImpl_AttackShouldFailWhenRegionNotInCache(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name          string
		regions       []snapshot.RegionState
		attackingID   string
		defendingID   string
		expectedError string
	}

	tests := []inputType{
		{
			"When attacking region not found in cache",
			[]snapshot.RegionState{
				{InternalID: 2, ID: "iceland", OwnerID: "gabriele", Troops: 5},
			},
			"greenland",
			"iceland",
			"unable to find attacking region in cache: region greenland not found in cached state",
		},
		{
			"When defending region not found in cache",
			[]snapshot.RegionState{
				{InternalID: 1, ID: "greenland", OwnerID: "giovanni", Troops: 5},
			},
			"greenland",
			"iceland",
			"unable to find defending region in cache: region iceland not found in cached state",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, _, service := setup(t)
			ctx := input()

			prev := &snapshot.CachedGameState{
				PublicSnapshot: &snapshot.GameSnapshot{
					Regions: test.regions,
				},
			}

			_, _, err := service.Perform(ctx, attack.Move{
				AttackingRegionID: test.attackingID,
				DefendingRegionID: test.defendingID,
				TroopsInSource:    5,
				TroopsInTarget:    5,
				AttackingTroops:   3,
			}, prev)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
}
