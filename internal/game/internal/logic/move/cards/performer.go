package cards

import (
	"fmt"

	cardsapi "github.com/go-risk-it/go-risk-it/internal/game/api/moves/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

func (s *service) Perform(
	ctx ctx.GameContext,
	move Move,
	prev *snapshot.CachedGameState,
) (*MoveResult, moveservice.MoveEffect, error) {
	var zero moveservice.MoveEffect

	cardIndex, err := buildCardIndex(ctx, prev)
	if err != nil {
		return nil, zero, err
	}

	extraDeployableTroops, playedCards, err := s.processCombinations(ctx, move, cardIndex)
	if err != nil {
		return nil, zero, err
	}

	regionTroopGrants := getRegionTroopGrants(ctx, prev, cardIndex, playedCards)

	result := &MoveResult{
		ExtraDeployableTroops: extraDeployableTroops,
		RegionTroopGrants:     regionTroopGrants,
	}

	// Collect the full CardState for each played card so they can be returned
	// to the available deck via DeckDelta.Returned.
	returnedCards := make([]snapshot.CardState, 0, len(playedCards))
	for _, cardID := range playedCards {
		returnedCards = append(returnedCards, cardIndex[cardID])
	}

	effect := moveservice.MoveEffect{
		CardDeltas: []moveservice.CardDelta{
			{
				PlayerUserID: ctx.UserID(),
				Lost:         playedCards,
			},
		},
		DeckDelta: moveservice.DeckDelta{
			Returned: returnedCards,
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	// Add region updates for each granted region (each gets DefaultTroopGrant bonus troops).
	// Use the cached prev state to get the current troop count so the MoveEffect carries
	// the correct absolute value (currentTroops + bonus), not just the bonus amount.
	cachedTroops := make(map[string]int64, 0)
	if prev != nil && prev.PublicSnapshot != nil {
		for _, r := range prev.PublicSnapshot.Regions {
			cachedTroops[r.ID] = r.Troops
		}
	}

	for _, grant := range regionTroopGrants {
		effect.RegionUpdates = append(effect.RegionUpdates, moveservice.RegionUpdate{
			RegionID:  grant.RegionExternalReference,
			NewOwner:  ctx.UserID(),
			NewTroops: cachedTroops[grant.RegionExternalReference] + DefaultTroopGrant,
		})
	}

	return result, effect, nil
}

func buildCardIndex(
	ctx ctx.GameContext,
	prev *snapshot.CachedGameState,
) (map[int64]snapshot.CardState, error) {
	if prev == nil || prev.PrivateSnapshots == nil {
		return nil, domainerrors.NewValidationError("cached game state is required for cards move")
	}

	playerPrivate, ok := prev.PrivateSnapshots[ctx.UserID()]
	if !ok {
		return nil, domainerrors.NewValidationError("no private snapshot for current player")
	}

	cardIndex := make(map[int64]snapshot.CardState, len(playerPrivate.Cards))
	for _, card := range playerPrivate.Cards {
		cardIndex[card.ID] = card
	}

	return cardIndex, nil
}

func (s *service) processCombinations(
	ctx ctx.GameContext,
	move Move,
	cardIndex map[int64]snapshot.CardState,
) (int64, []int64, error) {
	if len(move.Combinations) == 0 {
		return 0, nil, domainerrors.NewValidationError("no combinations provided")
	}

	if err := validateAllCardsDifferent(move); err != nil {
		return 0, nil, fmt.Errorf("validation failed: %w", err)
	}

	extraDeployableTroops := int64(0)
	playedCards := make([]int64, 0, len(move.Combinations)*3)

	for _, combination := range move.Combinations {
		if err := validateCombination(combination, cardIndex); err != nil {
			return 0, nil, fmt.Errorf("validation failed: %w", err)
		}

		combinationTroops, err := identifyCombination(combination, cardIndex)
		if err != nil {
			return 0, nil, fmt.Errorf("validation failed: %w", err)
		}

		extraDeployableTroops += combinationTroops

		playedCards = append(playedCards, combination.CardIDs...)
	}

	observe.SpanEvent(ctx, "processed_combinations",
		attribute.Int64("extra_troops", extraDeployableTroops),
	)

	return extraDeployableTroops, playedCards, nil
}

// RegionTroopGrant is a type alias to the canonical definition in game/api/moves/cards.
// This preserves backward compatibility for all existing consumers.
type RegionTroopGrant = cardsapi.RegionTroopGrant

func validateAllCardsDifferent(move Move) error {
	cardMap := make(map[int64]struct{})

	for _, combination := range move.Combinations {
		for _, cardID := range combination.CardIDs {
			if _, ok := cardMap[cardID]; ok {
				return domainerrors.NewValidationError("all cards must be different")
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

	// cardsPerCombination is the exact number of cards required in each combination.
	cardsPerCombination = 3
	// maxJollyPerCombination is the maximum jolly card factor allowed.
	// A combination value at or above this factor times JOLLY is invalid.
	maxJollyPerCombination = 2
)

func validateCombination(
	combination CardCombination,
	cardIndex map[int64]snapshot.CardState,
) error {
	if len(combination.CardIDs) != cardsPerCombination {
		return domainerrors.NewValidationError("combination must have exactly 3 cards")
	}

	// check if the cards are owned by this player
	for _, cardID := range combination.CardIDs {
		if _, ok := cardIndex[cardID]; !ok {
			return domainerrors.NewValidationError("player does not own one of the cards")
		}
	}

	return nil
}

func identifyCombination(
	combination CardCombination,
	cardIndex map[int64]snapshot.CardState,
) (int64, error) {
	type1 := cardIndex[combination.CardIDs[0]].Type
	type2 := cardIndex[combination.CardIDs[1]].Type
	type3 := cardIndex[combination.CardIDs[2]].Type

	combinationValue := getCardValue(type1) + getCardValue(type2) + getCardValue(type3)

	combinationToTroops := map[int64]int64{
		3 * ARTILLERY:                  4,
		3 * INFANTRY:                   6,
		3 * CAVALRY:                    8,
		ARTILLERY + INFANTRY + CAVALRY: 10,
		JOLLY + 2*ARTILLERY:            12,
		JOLLY + 2*INFANTRY:             12,
		JOLLY + 2*CAVALRY:              12,
	}

	if combinationValue >= maxJollyPerCombination*JOLLY {
		return 0, domainerrors.NewValidationError(
			"cannot use more than 2 jolly cards in a combination",
		)
	}

	if troops, ok := combinationToTroops[combinationValue]; ok {
		return troops, nil
	}

	return 0, domainerrors.NewValidationError("invalid combination")
}

func getCardValue(cardType snapshot.CardType) int64 {
	switch cardType {
	case snapshot.CardArtillery:
		return ARTILLERY
	case snapshot.CardInfantry:
		return INFANTRY
	case snapshot.CardCavalry:
		return CAVALRY
	default:
		return JOLLY
	}
}

func getRegionTroopGrants(
	ctx ctx.GameContext,
	prev *snapshot.CachedGameState,
	cardIndex map[int64]snapshot.CardState,
	playedCards []int64,
) []RegionTroopGrant {
	result := make([]RegionTroopGrant, 0)

	if prev == nil || prev.PublicSnapshot == nil {
		return result
	}

	playerRegions := getPlayerRegions(ctx, prev.PublicSnapshot.Regions)

	for _, cardID := range playedCards {
		card := cardIndex[cardID]
		if card.Region == "" {
			continue
		}

		for _, region := range playerRegions {
			if region.ID == card.Region {
				result = append(result, RegionTroopGrant{
					RegionID:                region.InternalID,
					RegionExternalReference: region.ID,
				})

				break
			}
		}
	}

	return result
}

func getPlayerRegions(
	ctx ctx.GameContext,
	regions []snapshot.RegionState,
) []snapshot.RegionState {
	result := make([]snapshot.RegionState, 0)

	for _, region := range regions {
		if region.OwnerID == ctx.UserID() {
			result = append(result, region)
		}
	}

	return result
}
