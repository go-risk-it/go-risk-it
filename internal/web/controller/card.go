package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/card"
)

type CardController interface {
	GetCardState(ctx ctx.GameContext) (messaging.CardState, error)
}

type CardControllerImpl struct {
	cardService card.Service
}

var _ CardController = (*CardControllerImpl)(nil)

func NewCardController(cardService card.Service) *CardControllerImpl {
	return &CardControllerImpl{
		cardService: cardService,
	}
}

func (c *CardControllerImpl) GetCardState(
	ctx ctx.GameContext,
) (messaging.CardState, error) {
	cards, err := c.cardService.GetCardsForPlayer(ctx)
	if err != nil {
		return messaging.CardState{}, fmt.Errorf("unable to get cards: %w", err)
	}

	return messaging.CardState{Cards: convertCards(cards)}, nil
}

func convertCards(cards []sqlc.GetCardsForPlayerRow) []messaging.Card {
	result := make([]messaging.Card, len(cards))
	for i, c := range cards {
		result[i] = convertCard(c)
	}

	return result
}

func convertCard(card sqlc.GetCardsForPlayerRow) messaging.Card {
	region := ""
	if card.Region.Valid {
		region = card.Region.String
	}

	return messaging.Card{
		Type:   convertCartType(card.CardType),
		Region: region,
	}
}

func convertCartType(sqlcCardType sqlc.CardType) messaging.CardType {
	switch sqlcCardType {
	case sqlc.CardTypeCAVALRY:
		return messaging.Cavalry
	case sqlc.CardTypeARTILLERY:
		return messaging.Infantry
	case sqlc.CardTypeINFANTRY:
		return messaging.Artillery
	case sqlc.CardTypeJOLLY:
		return messaging.Jolly
	default:
		panic("unknown card type")
	}
}
