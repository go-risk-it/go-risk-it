package cards

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
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

	regionTroopGrants, err := s.grantRegionTroops(ctx, querier, cardIndex, playedCards)
	if err != nil {
		return nil, fmt.Errorf("unable to grant region troops: %w", err)
	}

	return &MoveResult{
		ExtraDeployableTroops: extraDeployableTroops,
		RegionTroopGrants:     regionTroopGrants,
	}, nil
}

type RegionTroopGrant struct {
	RegionID                int64  `json:"regionId"`
	RegionExternalReference string `json:"regionExternalReference"`
}

func (s *ServiceImpl) grantRegionTroops(
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
		ctx.Log().Infow("no region troop grants")

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

	ctx.Log().Infof("granted bonus troops to %d regions", len(grantedRegionIds))

	return grants, nil
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

	combinationToTroops := map[int64]int64{
		3 * ARTILLERY:                  4,
		3 * INFANTRY:                   6,
		3 * CAVALRY:                    8,
		ARTILLERY + INFANTRY + CAVALRY: 10,
		JOLLY + 2*ARTILLERY:            12,
		JOLLY + 2*INFANTRY:             12,
		JOLLY + 2*CAVALRY:              12,
	}

	if combinationValue >= 2*JOLLY {
		return 0, errors.New("cannot use more than 2 jolly cards in a combination")
	}

	if troops, ok := combinationToTroops[combinationValue]; ok {
		return troops, nil
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

func (s *ServiceImpl) getRegionTroopGrants(
	ctx ctx.GameContext,
	querier db.Querier,
	cardIndex map[int64]sqlc.GetCardsForPlayerRow,
	playedCards []int64,
) ([]RegionTroopGrant, error) {
	result := make([]RegionTroopGrant, 0)

	regions, err := s.regionService.GetRegionsQ(ctx, querier)
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
