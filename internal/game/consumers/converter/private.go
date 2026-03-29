package converter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

// ConvertPrivateSnapshot transforms a PrivateSnapshot into pre-serialized WS
// messages for a single player: cardState and missionState.
//
// The missionResolver is called to fetch mission details (continent names,
// target player) that are not part of the snapshot. This keeps the converter
// testable — tests provide a stub, the signal handler provides the real
// implementation backed by MissionController.
func ConvertPrivateSnapshot(
	ctx context.Context,
	snap *snapshot.PrivateSnapshot,
	missionResolver MissionResolver,
) (*PrivateMessages, error) {
	cardMsg, err := buildCardStateMessage(snap.Cards)
	if err != nil {
		return nil, fmt.Errorf("building card state message: %w", err)
	}

	missionMsg, err := missionResolver(ctx, snap.MissionType, snap.MissionID)
	if err != nil {
		return nil, fmt.Errorf("resolving mission state: %w", err)
	}

	return &PrivateMessages{
		CardState:    cardMsg,
		MissionState: missionMsg,
	}, nil
}

func buildCardStateMessage(cards []sqlc.GetCardsForPlayerRow) (json.RawMessage, error) {
	convertedCards, err := convertCards(cards)
	if err != nil {
		return nil, err
	}

	return message.BuildMessage(message.CardState, messaging.CardState{Cards: convertedCards})
}

func convertCards(cards []sqlc.GetCardsForPlayerRow) ([]messaging.Card, error) {
	result := make([]messaging.Card, len(cards))
	for idx, card := range cards {
		card, err := convertCard(card)
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
