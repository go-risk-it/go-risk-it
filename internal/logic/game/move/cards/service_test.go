package cards_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/board"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/player"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/rand"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	*db.Querier,
	cards.Service,
	*region.Service,
) {
	t.Helper()
	querier := db.NewQuerier(t)
	boardService := board.NewService(t)
	phaseService := phase.NewService(t)
	playerService := player.NewService(t)
	regionService := region.NewService(t)
	rng := rand.NewRNG(t)
	service := cards.NewService(boardService, phaseService, playerService, regionService, rng)

	return querier, service, regionService
}

func input() ctx.GameContext {
	gameID := int64(1)
	userID := "giovanni"

	userContext := ctx.WithUserID(
		ctx.WithSpan(ctx.WithLog(context.Background(), zap.NewExample().Sugar()), noop.Span{}),
		userID,
	)

	return ctx.WithGameID(userContext, gameID)
}

func card(id int64, cardType sqlc.GameCardType) sqlc.GetCardsForPlayerRow {
	return sqlc.GetCardsForPlayerRow{
		ID:       id,
		CardType: cardType,
		Region: pgtype.Text{
			Valid: false,
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

			querier, service, _ := setup(t)
			ctx := input()

			querier.
				EXPECT().
				GetCardsForPlayer(ctx, sqlc.GetCardsForPlayerParams{
					ID:     ctx.GameID(),
					UserID: ctx.UserID(),
				}).
				Return([]sqlc.GetCardsForPlayerRow{
					card(1, sqlc.GameCardTypeARTILLERY),
					card(2, sqlc.GameCardTypeARTILLERY),
					card(3, sqlc.GameCardTypeARTILLERY),
					card(4, sqlc.GameCardTypeINFANTRY),
					card(5, sqlc.GameCardTypeINFANTRY),
					card(6, sqlc.GameCardTypeINFANTRY),
					card(7, sqlc.GameCardTypeCAVALRY),
					card(8, sqlc.GameCardTypeCAVALRY),
					card(9, sqlc.GameCardTypeCAVALRY),
					card(10, sqlc.GameCardTypeJOLLY),
					card(11, sqlc.GameCardTypeJOLLY),
				}, nil)

			_, err := service.PerformQ(ctx, querier, cards.Move{
				Combinations: test.combinations,
			})

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

			querier, service, regionService := setup(t)
			ctx := input()

			querier.
				EXPECT().
				GetCardsForPlayer(ctx, sqlc.GetCardsForPlayerParams{
					ID:     ctx.GameID(),
					UserID: ctx.UserID(),
				}).
				Return([]sqlc.GetCardsForPlayerRow{
					card(1, sqlc.GameCardTypeARTILLERY),
					card(2, sqlc.GameCardTypeARTILLERY),
					card(3, sqlc.GameCardTypeARTILLERY),
					card(4, sqlc.GameCardTypeINFANTRY),
					card(5, sqlc.GameCardTypeINFANTRY),
					card(6, sqlc.GameCardTypeINFANTRY),
					card(7, sqlc.GameCardTypeCAVALRY),
					card(8, sqlc.GameCardTypeCAVALRY),
					card(9, sqlc.GameCardTypeCAVALRY),
					card(10, sqlc.GameCardTypeJOLLY),
					card(11, sqlc.GameCardTypeJOLLY),
				}, nil)

			playedCards := make([]int64, 0)
			for _, combination := range test.combinations {
				playedCards = append(playedCards, combination.CardIDs...)
			}

			var regions []sqlc.GetRegionsByGameRow

			querier.
				EXPECT().
				UnlinkCardsFromOwner(ctx, playedCards).
				Return(nil)
			regionService.
				EXPECT().
				GetRegionsQ(ctx, querier).
				Return(regions, nil)

			extraTroops, _ := service.PerformQ(ctx, querier, cards.Move{
				Combinations: test.combinations,
			})

			require.Equal(t, test.expectedExtraTroops, extraTroops.ExtraDeployableTroops)
		})
	}
}
