package cards

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/board"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/region"
	"github.com/go-risk-it/go-risk-it/internal/game/rand"
)

const DefaultTroopGrant = 2

// minCardsForCombination is the minimum number of cards a player needs to attempt a combination.
const minCardsForCombination = 3

type CardCombination struct {
	CardIDs []int64 `json:"cardIds"`
}

type Move struct {
	Combinations []CardCombination `json:"combinations"`
}

type MoveResult struct {
	ExtraDeployableTroops int64              `json:"extraDeployableTroops"`
	RegionTroopGrants     []RegionTroopGrant `json:"regionTroopGrants"`
}

type Service interface {
	moveservice.Service[Move, *MoveResult]
	Draw(ctx ctx.GameContext, querier db.Querier) error
	NextPlayerHasValidCombination(ctx ctx.GameContext, querier db.Querier) (bool, error)
}

type service struct {
	boardService  board.Service
	phaseService  phase.Service
	playerService player.Service
	regionService region.Service
	rng           rand.RNG
}

var _ Service = (*service)(nil)

func NewService(
	boardService board.Service,
	phaseService phase.Service,
	playerService player.Service,
	regionService region.Service,
	rng rand.RNG,
) (Service, moveservice.Service[Move, *MoveResult]) {
	svc := &service{
		boardService:  boardService,
		phaseService:  phaseService,
		playerService: playerService,
		regionService: regionService,
		rng:           rng,
	}

	return svc, moveservice.NewTracedService[Move, *MoveResult](svc)
}

func (s *service) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeCARDS
}

func (s *service) Draw(ctx ctx.GameContext, querier db.Querier) error {
	slog.InfoContext(ctx, "drawing card")

	cards, err := querier.GetAvailableCards(ctx, ctx.GameID())
	if err != nil {
		return fmt.Errorf("failed to get available cards: %w", err)
	}

	if len(cards) == 0 {
		return errors.New("no cards available")
	}

	card := cards[s.rng.IntN(len(cards))]
	if err := querier.DrawCard(ctx, sqlc.DrawCardParams{
		ID:     card.ID,
		UserID: ctx.UserID(),
		GameID: ctx.GameID(),
	}); err != nil {
		return fmt.Errorf("failed to draw card: %w", err)
	}

	slog.InfoContext(ctx, "card drawn", "card", card.ID)

	return nil
}

func (s *service) NextPlayerHasValidCombination(
	ctx ctx.GameContext,
	querier db.Querier,
) (bool, error) {
	slog.DebugContext(ctx, "checking if the player has a valid card combination")

	nextPlayer, err := s.playerService.GetNextPlayer(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get player: %w", err)
	}

	nextPlayerCards, err := querier.GetCardsForPlayer(ctx, sqlc.GetCardsForPlayerParams{
		ID:     ctx.GameID(),
		UserID: nextPlayer.UserID,
	})
	if err != nil {
		return false, fmt.Errorf("unable to get cards for player: %w", err)
	}

	slog.DebugContext(ctx, "player has cards",
		"count", len(nextPlayerCards), "cards", nextPlayerCards)

	if len(nextPlayerCards) < minCardsForCombination {
		return false, nil
	}

	cardIndex := make(map[int64]sqlc.GetCardsForPlayerRow)
	for _, card := range nextPlayerCards {
		cardIndex[card.ID] = card
	}

	// Try each possible combination of 3 cards
	for i := range len(nextPlayerCards) - 2 {
		for j := i + 1; j < len(nextPlayerCards)-1; j++ {
			for k := j + 1; k < len(nextPlayerCards); k++ {
				combination := CardCombination{
					CardIDs: []int64{
						nextPlayerCards[i].ID,
						nextPlayerCards[j].ID,
						nextPlayerCards[k].ID,
					},
				}

				slog.DebugContext(ctx, "checking combination", "combination", combination)

				if _, err := identifyCombination(combination, cardIndex); err == nil {
					slog.DebugContext(ctx, "player has a valid combination",
						"combination", combination)

					return true, nil
				}
			}
		}
	}

	slog.DebugContext(ctx, "player has no valid combination")

	return false, nil
}
