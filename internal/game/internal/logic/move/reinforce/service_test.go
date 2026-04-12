package reinforce_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/reinforce"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/move/cards"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (*board.Service, *cards.Service, reinforce.Service) {
	t.Helper()

	boardService := board.NewService(t)
	cardsService := cards.NewService(t)

	service, _ := reinforce.NewService(boardService, cardsService)

	return boardService, cardsService, service
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

func cachedState(
	sourceID string,
	sourceOwner string,
	sourceTroops int64,
	targetID string,
	targetOwner string,
	targetTroops int64,
) *snapshot.CachedGameState {
	return &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Regions: []snapshot.RegionState{
				{
					InternalID: 1,
					ID:         sourceID,
					OwnerID:    sourceOwner,
					Troops:     sourceTroops,
				},
				{
					InternalID: 2,
					ID:         targetID,
					OwnerID:    targetOwner,
					Troops:     targetTroops,
				},
			},
		},
	}
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

			boardService, _, service := setup(t)
			ctx := input()

			prev := cachedState(
				test.sourceRegionID, test.sourceOwner, test.troopsInSource,
				test.targetRegionID, test.targetOwner, test.troopsInTarget,
			)

			if !test.canReach {
				boardService.
					EXPECT().
					CanPlayerReachWithRegions(
						ctx,
						test.sourceRegionID,
						test.targetRegionID,
						prev.PublicSnapshot.Regions,
					).
					Return(false, nil)
			}

			_, _, err := service.Perform(ctx, reinforce.Move{
				SourceRegionID: test.sourceRegionID,
				TargetRegionID: test.targetRegionID,
				TroopsInSource: test.declaredTroopsInSrc,
				TroopsInTarget: test.declaredTroopsInTgt,
				MovingTroops:   test.movingTroops,
			}, prev)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestServiceImpl_Perform_ShouldMoveTroops(t *testing.T) {
	t.Parallel()

	boardService, _, service := setup(t)
	ctx := input()

	prev := cachedState(
		"greenland", "giovanni", 5,
		"iceland", "giovanni", 3,
	)

	boardService.
		EXPECT().
		CanPlayerReachWithRegions(
			ctx,
			"greenland",
			"iceland",
			prev.PublicSnapshot.Regions,
		).
		Return(true, nil)

	_, effect, err := service.Perform(ctx, reinforce.Move{
		SourceRegionID: "greenland",
		TargetRegionID: "iceland",
		TroopsInSource: 5,
		TroopsInTarget: 3,
		MovingTroops:   3,
	}, prev)

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

func TestAdvance_ConqueredInTurn_DrawsCardAndPersistsViaDelta(t *testing.T) {
	t.Parallel()

	_, cardsService, service := setup(t)
	ctx := input()

	deck := []snapshot.CardState{
		{ID: 100, Type: snapshot.CardArtillery, Region: "brazil"},
		{ID: 200, Type: snapshot.CardInfantry, Region: "iceland"},
	}

	drawnCard := snapshot.CardState{ID: 200, Type: snapshot.CardInfantry, Region: "iceland"}

	advCtx := moveservice.AdvanceContext{
		ConqueredInTurn: true,
		AvailableDeck:   deck,
		Turn:            0,
		Players: []snapshot.PlayerState{
			{UserID: "giovanni", Index: 0, Status: snapshot.PlayerAlive},
		},
	}

	// Expect Draw to be called with the deck.
	cardsService.EXPECT().Draw(deck).Return(drawnCard, nil)

	// Advancing to CARDS phase.
	// Cards service Advance is NOT called when going to CARDS (only to DEPLOY).
	effect, err := service.Advance(ctx, sqlc.GamePhaseTypeCARDS, struct{}{}, advCtx)

	require.NoError(t, err)
	require.True(t, effect.TurnEnded)

	// Verify CardDeltas: player gained the drawn card.
	require.Len(t, effect.CardDeltas, 1)
	require.Equal(t, moveservice.CardDelta{
		PlayerUserID: "giovanni",
		Gained:       []snapshot.CardState{drawnCard},
	}, effect.CardDeltas[0])

	// Verify DeckDelta.Drawn records the drawn card ID.
	require.Equal(t, []int64{200}, effect.DeckDelta.Drawn)
}

func TestAdvance_NoConquest_NoDraw(t *testing.T) {
	t.Parallel()

	_, _, service := setup(t)
	ctx := input()

	advCtx := moveservice.AdvanceContext{
		ConqueredInTurn: false,
		AvailableDeck:   []snapshot.CardState{{ID: 100}},
		Turn:            0,
		Players: []snapshot.PlayerState{
			{UserID: "giovanni", Index: 0, Status: snapshot.PlayerAlive},
		},
	}

	// No Draw call expected. Advancing to CARDS phase.
	effect, err := service.Advance(ctx, sqlc.GamePhaseTypeCARDS, struct{}{}, advCtx)

	require.NoError(t, err)
	require.True(t, effect.TurnEnded)
	require.Empty(t, effect.CardDeltas)
	require.Empty(t, effect.DeckDelta.Drawn)
	require.Empty(t, effect.DeckDelta.Returned)
}
