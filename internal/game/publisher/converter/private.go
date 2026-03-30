package converter

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/snapshot"
)

// ConvertPrivateSnapshot transforms a PrivateSnapshot into typed DTOs for a
// single player: cardState and missionState.
//
// The missionResolver is called to fetch mission details (continent names,
// target player) that are not part of the snapshot. This keeps the converter
// testable -- tests provide a stub, the signal handler provides the real
// implementation backed by MissionController.
func ConvertPrivateSnapshot(
	ctx context.Context,
	snap *snapshot.PrivateSnapshot,
	missionResolver MissionResolver,
) (*PrivateMessages, error) {
	cardState, err := buildCardState(snap.Cards)
	if err != nil {
		return nil, fmt.Errorf("building card state: %w", err)
	}

	missionState, err := missionResolver(ctx, snap.MissionType, snap.MissionID)
	if err != nil {
		return nil, fmt.Errorf("resolving mission state: %w", err)
	}

	return &PrivateMessages{
		CardState:    cardState,
		MissionState: missionState,
	}, nil
}

func buildCardState(cards []sqlc.GetCardsForPlayerRow) (messaging.CardState, error) {
	convertedCards, err := convertCards(cards)
	if err != nil {
		return messaging.CardState{}, err
	}

	return messaging.CardState{Cards: convertedCards}, nil
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
