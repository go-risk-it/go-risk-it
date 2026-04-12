package cards_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/cards"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/rand"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (*rand.RNG, cards.Service) {
	t.Helper()

	rng := rand.NewRNG(t)
	service, _ := cards.NewService(rng)

	return rng, service
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

func card(id int64, cardType snapshot.CardType) snapshot.CardState {
	return snapshot.CardState{
		ID:     id,
		Type:   cardType,
		Region: "",
	}
}

func defaultCards() []snapshot.CardState {
	return []snapshot.CardState{
		card(1, snapshot.CardArtillery),
		card(2, snapshot.CardArtillery),
		card(3, snapshot.CardArtillery),
		card(4, snapshot.CardInfantry),
		card(5, snapshot.CardInfantry),
		card(6, snapshot.CardInfantry),
		card(7, snapshot.CardCavalry),
		card(8, snapshot.CardCavalry),
		card(9, snapshot.CardCavalry),
		card(10, snapshot.CardJolly),
		card(11, snapshot.CardJolly),
	}
}

func cachedState(userID string, playerCards []snapshot.CardState) *snapshot.CachedGameState {
	return &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Regions: []snapshot.RegionState{},
		},
		PrivateSnapshots: map[string]*snapshot.PlayerPrivate{
			userID: {
				Cards: playerCards,
			},
		},
	}
}

func TestServiceImpl_InvalidCombinations(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name          string
		combinations  []cards.CardCombination
		expectedError string
	}

	tests := []inputType{
		{
			name:          "Use no cards",
			combinations:  []cards.CardCombination{},
			expectedError: "no combinations provided",
		},
		{
			name: "Empty combination",
			combinations: []cards.CardCombination{
				{},
			},
			expectedError: "validation failed: combination must have exactly 3 cards",
		},
		{
			name:          "Use a single card",
			combinations:  []cards.CardCombination{{CardIDs: []int64{2}}},
			expectedError: "validation failed: combination must have exactly 3 cards",
		},
		{
			name:          "One card is used twice in the same combination",
			combinations:  []cards.CardCombination{{CardIDs: []int64{1, 1, 2}}},
			expectedError: "validation failed: all cards must be different",
		},
		{
			name: "One card is used twice in different combinations",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{1, 2, 3}},
				{CardIDs: []int64{1, 4, 5}},
			},
			expectedError: "validation failed: all cards must be different",
		},
		{
			name: "Use a card that is not owned",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{1, 2, 3}},
				{CardIDs: []int64{4, 5, 17}},
			},
			expectedError: "validation failed: player does not own one of the cards",
		},
		{
			name: "Two artillery cards and one infantry card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{1, 2, 4}},
			},
			expectedError: "validation failed: invalid combination",
		},
		{
			name: "Two artillery cards and one cavalry card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{1, 2, 7}},
			},
			expectedError: "validation failed: invalid combination",
		},
		{
			name: "Two infantry cards and one artillery card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{4, 5, 1}},
			},
			expectedError: "validation failed: invalid combination",
		},
		{
			name: "Two infantry cards and one cavalry card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{4, 5, 7}},
			},
			expectedError: "validation failed: invalid combination",
		},
		{
			name: "Two cavalry cards and one artillery card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{7, 8, 1}},
			},
			expectedError: "validation failed: invalid combination",
		},
		{
			name: "Two cavalry cards and one infantry card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{7, 8, 4}},
			},
			expectedError: "validation failed: invalid combination",
		},
		{
			name: "One jolly card, one artillery card and one infantry card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{10, 1, 4}},
			},
			expectedError: "validation failed: invalid combination",
		},
		{
			name: "One jolly card, one artillery card and one cavalry card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{10, 1, 7}},
			},
			expectedError: "validation failed: invalid combination",
		},
		{
			name: "One jolly card, one infantry card and one cavalry",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{10, 4, 7}},
			},
			expectedError: "validation failed: invalid combination",
		},
		{
			name: "Two jolly cards and one artillery card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{10, 11, 1}},
			},
			expectedError: "validation failed: cannot use more than 2 jolly cards in a combination",
		},
		{
			name: "Two jolly cards and one infantry card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{10, 11, 4}},
			},
			expectedError: "validation failed: cannot use more than 2 jolly cards in a combination",
		},
		{
			name: "Two jolly cards and one cavalry card",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{10, 11, 7}},
			},
			expectedError: "validation failed: cannot use more than 2 jolly cards in a combination",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, service := setup(t)
			ctx := input()
			prev := cachedState(ctx.UserID(), defaultCards())

			_, _, err := service.Perform(ctx, cards.Move{
				Combinations: test.combinations,
			}, prev)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestServiceImpl_ValidCombinations(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name                string
		combinations        []cards.CardCombination
		expectedExtraTroops int64
	}

	tests := []inputType{
		{
			name:                "Artillery combination",
			combinations:        []cards.CardCombination{{CardIDs: []int64{1, 2, 3}}},
			expectedExtraTroops: 4,
		},
		{
			name:                "Infantry combination",
			combinations:        []cards.CardCombination{{CardIDs: []int64{4, 5, 6}}},
			expectedExtraTroops: 6,
		},
		{
			name:                "Cavalry combination",
			combinations:        []cards.CardCombination{{CardIDs: []int64{7, 8, 9}}},
			expectedExtraTroops: 8,
		},
		{
			name:                "One of each type",
			combinations:        []cards.CardCombination{{CardIDs: []int64{1, 4, 7}}},
			expectedExtraTroops: 10,
		},
		{
			name:                "Two artillery cards and one jolly card",
			combinations:        []cards.CardCombination{{CardIDs: []int64{1, 2, 10}}},
			expectedExtraTroops: 12,
		},
		{
			name:                "Two infantry cards and one jolly card",
			combinations:        []cards.CardCombination{{CardIDs: []int64{4, 5, 10}}},
			expectedExtraTroops: 12,
		},
		{
			name:                "Two cavalry cards and one jolly card",
			combinations:        []cards.CardCombination{{CardIDs: []int64{7, 8, 10}}},
			expectedExtraTroops: 12,
		},
		{
			name: "Three combinations",
			combinations: []cards.CardCombination{
				{CardIDs: []int64{1, 2, 3}},
				{CardIDs: []int64{4, 5, 10}},
				{CardIDs: []int64{7, 8, 11}},
			},
			expectedExtraTroops: 28,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, service := setup(t)
			ctx := input()
			prev := cachedState(ctx.UserID(), defaultCards())

			playedCards := make([]int64, 0)
			for _, combination := range test.combinations {
				playedCards = append(playedCards, combination.CardIDs...)
			}

			result, effect, err := service.Perform(ctx, cards.Move{
				Combinations: test.combinations,
			}, prev)

			require.NoError(t, err)
			require.Equal(t, test.expectedExtraTroops, result.ExtraDeployableTroops)

			// Verify MoveEffect
			require.Len(t, effect.CardDeltas, 1)
			require.Equal(t, moveservice.CardDelta{
				PlayerUserID: "giovanni",
				Lost:         playedCards,
			}, effect.CardDeltas[0])
			require.Empty(t, effect.RegionUpdates)
			require.Equal(t, snapshot.EmptyPhaseState{}, effect.UpdatedPhase)

			// Verify DeckDelta.Returned contains the played cards' full CardState.
			allCards := defaultCards()
			expectedReturned := make([]snapshot.CardState, 0, len(playedCards))
			cardsByID := make(map[int64]snapshot.CardState, len(allCards))
			for _, c := range allCards {
				cardsByID[c.ID] = c
			}
			for _, id := range playedCards {
				expectedReturned = append(expectedReturned, cardsByID[id])
			}
			require.Equal(t, expectedReturned, effect.DeckDelta.Returned)
		})
	}
}

func TestServiceImpl_ShouldGrantRegionTroopsWhenPlayedCardMatchesOwnedRegion(t *testing.T) {
	t.Parallel()

	_, service := setup(t)
	ctx := input()

	// Player "giovanni" owns "iceland" and has a card with Region="iceland".
	playerCards := []snapshot.CardState{
		{ID: 1, Type: snapshot.CardArtillery, Region: "iceland"},
		{ID: 2, Type: snapshot.CardArtillery, Region: "brazil"},
		{ID: 3, Type: snapshot.CardArtillery, Region: ""},
	}

	prev := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Regions: []snapshot.RegionState{
				{InternalID: 42, ID: "iceland", OwnerID: "giovanni", Troops: 3},
				{InternalID: 43, ID: "brazil", OwnerID: "gabriele", Troops: 5},
			},
		},
		PrivateSnapshots: map[string]*snapshot.PlayerPrivate{
			"giovanni": {
				Cards: playerCards,
			},
		},
	}

	result, effect, err := service.Perform(ctx, cards.Move{
		Combinations: []cards.CardCombination{
			{CardIDs: []int64{1, 2, 3}},
		},
	}, prev)

	require.NoError(t, err)
	require.Equal(t, int64(4), result.ExtraDeployableTroops) // 3 artillery = 4 troops

	// Region troop grants: only "iceland" matches (owned by giovanni + card has region)
	require.Len(t, result.RegionTroopGrants, 1)
	require.Equal(t, int64(42), result.RegionTroopGrants[0].RegionID)
	require.Equal(t, "iceland", result.RegionTroopGrants[0].RegionExternalReference)

	// Verify MoveEffect includes region update for the bonus
	require.Len(t, effect.RegionUpdates, 1)
	require.Equal(t, moveservice.RegionUpdate{
		RegionID:  "iceland",
		NewOwner:  "giovanni",
		NewTroops: 3 + cards.DefaultTroopGrant, // 3 existing + 2 bonus
	}, effect.RegionUpdates[0])
}

func TestDraw_SelectsFromDeck(t *testing.T) {
	t.Parallel()

	rng, service := setup(t)

	deck := []snapshot.CardState{
		{ID: 10, Type: snapshot.CardArtillery, Region: "brazil"},
		{ID: 20, Type: snapshot.CardInfantry, Region: "iceland"},
		{ID: 30, Type: snapshot.CardCavalry, Region: "siam"},
	}

	// RNG returns index 1 → expect card with ID 20.
	rng.EXPECT().IntN(3).Return(1)

	drawn, err := service.Draw(deck)
	require.NoError(t, err)
	require.Equal(t, snapshot.CardState{
		ID:     20,
		Type:   snapshot.CardInfantry,
		Region: "iceland",
	}, drawn)
}

func TestDraw_EmptyDeck_ReturnsError(t *testing.T) {
	t.Parallel()

	_, service := setup(t)

	_, err := service.Draw(nil)
	require.Error(t, err)

	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	require.Equal(t, domainerrors.CategoryValidation, domainErr.Category())
	require.EqualError(t, err, "no cards available in deck")
}

func TestDraw_SingleCard_ReturnsThatCard(t *testing.T) {
	t.Parallel()

	rng, service := setup(t)

	deck := []snapshot.CardState{
		{ID: 42, Type: snapshot.CardJolly, Region: ""},
	}

	rng.EXPECT().IntN(1).Return(0)

	drawn, err := service.Draw(deck)
	require.NoError(t, err)
	require.Equal(t, deck[0], drawn)
}
