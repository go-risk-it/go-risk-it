package cards

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

func (s *service) Perform(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
) (*MoveResult, error) {
	slog.DebugContext(ctx, "performing cards move", "move", move)

	cardIndex, err := s.buildCardIndex(ctx, querier)
	if err != nil {
		return nil, err
	}

	extraDeployableTroops, playedCards, err := s.processCombinations(ctx, move, cardIndex)
	if err != nil {
		return nil, err
	}

	err = querier.UnlinkCardsFromOwner(ctx, playedCards)
	if err != nil {
		return nil, fmt.Errorf("unable to unlink cards from owner: %w", err)
	}

	regionTroopGrants, err := s.grantRegionTroops(ctx, querier, cardIndex, playedCards)
	if err != nil {
		return nil, fmt.Errorf("unable to grant region troops: %w", err)
	}

	result := &MoveResult{
		ExtraDeployableTroops: extraDeployableTroops,
		RegionTroopGrants:     regionTroopGrants,
	}

	return result, nil
}

func (s *service) buildCardIndex(
	ctx ctx.GameContext,
	querier db.Querier,
) (map[int64]sqlc.GetCardsForPlayerRow, error) {
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

	return cardIndex, nil
}

func (s *service) processCombinations(
	ctx ctx.GameContext,
	move Move,
	cardIndex map[int64]sqlc.GetCardsForPlayerRow,
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

	slog.DebugContext(ctx, "processed combinations", "extraTroops", extraDeployableTroops)

	return extraDeployableTroops, playedCards, nil
}

type RegionTroopGrant struct {
	RegionID                int64  `json:"regionId"`
	RegionExternalReference string `json:"regionExternalReference"`
}

func (s *service) grantRegionTroops(
	ctx ctx.GameContext,
	querier db.Querier,
	cardIndex map[int64]sqlc.GetCardsForPlayerRow,
	playedCards []int64,
) ([]RegionTroopGrant, error) {
	grants, err := s.getRegionTroopGrants(ctx, querier, cardIndex, playedCards)
	if err != nil {
		return nil, fmt.Errorf("failed to get region troop grants: %w", err)
	}

	if len(grants) == 0 {
		slog.DebugContext(ctx, "no region troop grants")

		return nil, nil
	}

	grantedRegionIds := make([]int64, 0)
	for _, grant := range grants {
		grantedRegionIds = append(grantedRegionIds, grant.RegionID)
	}

	if err := querier.GrantRegionTroops(ctx, sqlc.GrantRegionTroopsParams{
		Regions: grantedRegionIds,
		Troops:  DefaultTroopGrant,
	}); err != nil {
		return nil, fmt.Errorf("failed to grant region troops: %w", err)
	}

	slog.DebugContext(ctx, "granted bonus troops to regions", "count", len(grantedRegionIds))

	return grants, nil
}

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
	cardIndex map[int64]sqlc.GetCardsForPlayerRow,
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
	cardIndex map[int64]sqlc.GetCardsForPlayerRow,
) (int64, error) {
	type1 := cardIndex[combination.CardIDs[0]].CardType
	type2 := cardIndex[combination.CardIDs[1]].CardType
	type3 := cardIndex[combination.CardIDs[2]].CardType

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

func getCardValue(cardType sqlc.GameCardType) int64 {
	switch cardType {
	case sqlc.GameCardTypeARTILLERY:
		return ARTILLERY
	case sqlc.GameCardTypeINFANTRY:
		return INFANTRY
	case sqlc.GameCardTypeCAVALRY:
		return CAVALRY
	default:
		return JOLLY
	}
}

func (s *service) getRegionTroopGrants(
	ctx ctx.GameContext,
	querier db.Querier,
	cardIndex map[int64]sqlc.GetCardsForPlayerRow,
	playedCards []int64,
) ([]RegionTroopGrant, error) {
	result := make([]RegionTroopGrant, 0)

	regions, err := s.regionService.GetRegionsWithQuerier(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}

	playerRegions := getPlayerRegionsWithID(ctx, regions)

	for _, cardID := range playedCards {
		card := cardIndex[cardID]
		if !card.Region.Valid {
			continue
		}

		index := slices.IndexFunc(playerRegions, func(regionRow sqlc.GetRegionsByGameRow) bool {
			return regionRow.ExternalReference == card.Region.String
		})
		if index == -1 {
			continue
		}

		region := playerRegions[index]
		result = append(result, RegionTroopGrant{
			RegionID:                region.ID,
			RegionExternalReference: region.ExternalReference,
		})
	}

	return result, nil
}

func getPlayerRegionsWithID(
	ctx ctx.GameContext,
	regions []sqlc.GetRegionsByGameRow,
) []sqlc.GetRegionsByGameRow {
	result := make([]sqlc.GetRegionsByGameRow, 0)

	for _, region := range regions {
		if region.UserID == ctx.UserID() {
			result = append(result, region)
		}
	}

	return result
}
