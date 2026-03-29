package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/card"
)

type CardController struct {
	cardService card.Service
}

func NewCardController(cardService card.Service) *CardController {
	return &CardController{
		cardService: cardService,
	}
}

func (c *CardController) GetCardState(
	ctx ctx.GameContext,
) (messaging.CardState, error) {
	cards, err := c.cardService.GetCardsForPlayer(ctx)
	if err != nil {
		return messaging.CardState{}, fmt.Errorf("unable to get cards: %w", err)
	}

	convertedCards, err := convertCards(cards)
	if err != nil {
		return messaging.CardState{}, fmt.Errorf("unable to convert cards: %w", err)
	}

	return messaging.CardState{Cards: convertedCards}, nil
}

func convertCards(cards []sqlc.GetCardsForPlayerRow) ([]messaging.Card, error) {
	result := make([]messaging.Card, len(cards))
	for idx, c := range cards {
		card, err := convertCard(c)
		if err != nil {
			return nil, err
		}

		result[idx] = card
	}

	return result, nil
}

func convertCard(card sqlc.GetCardsForPlayerRow) (messaging.Card, error) {
	region := ""
	if card.Region.Valid {
		region = card.Region.String
	}

	cardType, err := convertCardType(card.CardType)
	if err != nil {
		return messaging.Card{}, err
	}

	return messaging.Card{
		ID:     card.ID,
		Type:   cardType,
		Region: region,
	}, nil
}

func convertCardType(sqlcCardType sqlc.GameCardType) (messaging.CardType, error) {
	switch sqlcCardType {
	case sqlc.GameCardTypeCAVALRY:
		return messaging.Cavalry, nil
	case sqlc.GameCardTypeARTILLERY:
		return messaging.Artillery, nil
	case sqlc.GameCardTypeINFANTRY:
		return messaging.Infantry, nil
	case sqlc.GameCardTypeJOLLY:
		return messaging.Jolly, nil
	default:
		return "", fmt.Errorf("unknown card type: %s", sqlcCardType)
	}
}
