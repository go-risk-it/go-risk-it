package cards

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/rand"
)

const DefaultTroopGrant = 2

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
	service.Service[Move, *MoveResult]
	Draw(ctx ctx.GameContext, querier db.Querier) error
	NextPlayerHasValidCombinationQ(ctx ctx.GameContext, querier db.Querier) (bool, error)
}

type ServiceImpl struct {
	boardService  board.Service
	phaseService  phase.Service
	playerService player.Service
	regionService region.Service
	rng           rand.RNG
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	boardService board.Service,
	phaseService phase.Service,
	playerService player.Service,
	regionService region.Service,
	rng rand.RNG,
) *ServiceImpl {
	return &ServiceImpl{
		boardService:  boardService,
		phaseService:  phaseService,
		playerService: playerService,
		regionService: regionService,
		rng:           rng,
	}
}

func (s *ServiceImpl) PhaseType() sqlc.PhaseType {
	return sqlc.PhaseTypeCARDS
}

func (s *ServiceImpl) Draw(ctx ctx.GameContext, querier db.Querier) error {
	ctx.Log().Infow("drawing card")

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

	ctx.Log().Infow("card drawn", "card", card.ID)

	return nil
}

func (s *ServiceImpl) NextPlayerHasValidCombinationQ(
	ctx ctx.GameContext,
	querier db.Querier,
) (bool, error) {
	ctx.Log().Infow("checking if the player has a valid card combination")

	nextPlayer, err := s.playerService.GetNextPlayerQ(ctx, querier)
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

	ctx.Log().Debugf("player has %d cards: %v", len(nextPlayerCards), nextPlayerCards)

	if len(nextPlayerCards) < 3 {
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

				ctx.Log().Debugw("checking combination", "combination", combination)

				if _, err := identifyCombination(combination, cardIndex); err == nil {
					ctx.Log().Infow("player has a valid combination", "combination", combination)

					return true, nil
				}
			}
		}
	}

	ctx.Log().Infow("player has no valid combination")

	return false, nil
}
