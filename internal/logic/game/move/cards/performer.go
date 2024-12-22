package cards

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) PerformQ(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
) (*MoveResult, error) {
	ctx.Log().Infow("performing cards move", "move", move)

	extraDeployableTroops := int64(0)

	thisPlayerCards, err := querier.GetCardsForPlayer(ctx, sqlc.GetCardsForPlayerParams{
		ID:     ctx.GameID(),
		UserID: ctx.UserID(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get cards for player: %w", err)
	}

	cardIndex := make(map[int64]sqlc.GetCardsForPlayerRow)
	for _, card := range thisPlayerCards {
		cardIndex[card.ID] = card
	}

	if len(move.Combinations) == 0 {
		return nil, errors.New("no combinations provided")
	}

	if err := validateAllCardsDifferent(move); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	playedCards := make([]int64, 0, len(move.Combinations)*3)

	for _, combination := range move.Combinations {
		if err := validateCombination(combination, cardIndex); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		combinationTroops, err := identifyCombination(combination, cardIndex)
		if err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		extraDeployableTroops += combinationTroops

		playedCards = append(playedCards, combination.CardIDs...)
	}

	err = querier.UnlinkCardsFromOwner(ctx, playedCards)
	if err != nil {
		return nil, fmt.Errorf("unable to unlink cards from owner: %w", err)
	}

	return &MoveResult{
		ExtraDeployableTroops: extraDeployableTroops,
	}, nil
}

func validateAllCardsDifferent(move Move) error {
	cardMap := make(map[int64]struct{})

	for _, combination := range move.Combinations {
		for _, cardID := range combination.CardIDs {
			if _, ok := cardMap[cardID]; ok {
				return errors.New("all cards must be different")
			}

			cardMap[cardID] = struct{}{}
		}
	}

	return nil
}

const (
	ARTILLERY = 1
	INFANTRY  = 10
	CAVALRY   = 100
	JOLLY     = 1000
)

func validateCombination(
	combination CardCombination,
	cardIndex map[int64]sqlc.GetCardsForPlayerRow,
) error {
	if len(combination.CardIDs) != 3 {
		return errors.New("combination must have exactly 3 cards")
	}

	// check if the cards are owned by this player
	for _, cardID := range combination.CardIDs {
		if _, ok := cardIndex[cardID]; !ok {
			return errors.New("player does not own one of the cards")
		}
	}

	return nil
}

func identifyCombination(
	combination CardCombination,
	cardIndex map[int64]sqlc.GetCardsForPlayerRow,
) (int64, error) {
	type1 := cardIndex[combination.CardIDs[0]].CardType
	type2 := cardIndex[combination.CardIDs[1]].CardType
	type3 := cardIndex[combination.CardIDs[2]].CardType

	combinationValue := getCardValue(type1) + getCardValue(type2) + getCardValue(type3)

	if combinationValue >= 2*JOLLY {
		return 0, errors.New("cannot use more than 2 jolly cards in a combination")
	}

	if combinationValue == 3*ARTILLERY {
		return 4, nil
	}

	if combinationValue == 3*INFANTRY {
		return 6, nil
	}

	if combinationValue == 3*CAVALRY {
		return 8, nil
	}

	if combinationValue == ARTILLERY+INFANTRY+CAVALRY {
		return 10, nil
	}

	if combinationValue == JOLLY+2*ARTILLERY ||
		combinationValue == JOLLY+2*INFANTRY ||
		combinationValue == JOLLY+2*CAVALRY {
		return 12, nil
	}

	return 0, errors.New("invalid combination")
}

func getCardValue(cardType sqlc.CardType) int64 {
	switch cardType {
	case sqlc.CardTypeARTILLERY:
		return ARTILLERY
	case sqlc.CardTypeINFANTRY:
		return INFANTRY
	case sqlc.CardTypeCAVALRY:
		return CAVALRY
	default:
		return JOLLY
	}
}
