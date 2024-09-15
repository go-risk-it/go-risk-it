package attack

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) PerformQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	move Move,
) (*MoveResult, error) {
	ctx.Log().Infow("performing attack move", "move", move)

	attackingRegion, err := s.regionService.GetRegionQ(ctx, querier, move.AttackingRegionID)
	if err != nil {
		return nil, fmt.Errorf("unable to get attacking region: %w", err)
	}

	defendingRegion, err := s.regionService.GetRegionQ(ctx, querier, move.DefendingRegionID)
	if err != nil {
		return nil, fmt.Errorf("unable to get defending region: %w", err)
	}

	if err := s.validate(ctx, attackingRegion, defendingRegion, move); err != nil {
		ctx.Log().Infow("validation failed", "error", err)

		return nil, fmt.Errorf("validation failed: %w", err)
	}

	casualties, err := s.perform(ctx, querier, attackingRegion, defendingRegion, move)
	if err != nil {
		return nil, fmt.Errorf("unable to perform attack move: %w", err)
	}

	ctx.Log().Infow("attack executed successfully")

	return &MoveResult{
		AttackingRegionID: move.AttackingRegionID,
		DefendingRegionID: move.DefendingRegionID,
		ConqueringTroops:  move.AttackingTroops - casualties.attacking,
	}, nil
}

func (s *ServiceImpl) perform(
	ctx ctx.MoveContext,
	querier db.Querier,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) (*casualties, error) {
	attackDices := s.diceService.RollAttackingDices(int(move.AttackingTroops))
	defenseDices := s.diceService.RollDefendingDices(int(min(defendingRegion.Troops, 3)))

	ctx.Log().Infow("rolled dices", "attack", attackDices, "defense", defenseDices)

	casualties := computeCasualties(ctx, attackDices, defenseDices)

	ctx.Log().Infow("updating region troops")

	if err := s.regionService.UpdateTroopsInRegionQ(
		ctx,
		querier,
		attackingRegion,
		-casualties.attacking,
	); err != nil {
		return nil, fmt.Errorf("failed to decrease troops in attacking region: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegionQ(
		ctx,
		querier,
		defendingRegion,
		-casualties.defending,
	); err != nil {
		return nil, fmt.Errorf("failed to decrease troops in defending region: %w", err)
	}

	return casualties, nil
}

type casualties struct {
	attacking int64
	defending int64
}

func computeCasualties(ctx ctx.MoveContext, attackDices, defenseDices []int) *casualties {
	casualties := &casualties{}

	slices.SortFunc(attackDices, descending)
	slices.SortFunc(defenseDices, descending)

	matches := min(len(attackDices), len(defenseDices))
	for i := range matches {
		if attackDices[i] > defenseDices[i] {
			casualties.defending++
		} else {
			casualties.attacking++
		}
	}

	ctx.Log().Infow(
		"casualties",
		"attacking",
		casualties.attacking,
		"defending",
		casualties.defending)

	return casualties
}

func descending(a, b int) int {
	return b - a
}

func (s *ServiceImpl) validate(
	ctx ctx.MoveContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	ctx.Log().Infow("validating attack move", "move", move)

	if err := checkRegionOwnership(ctx, attackingRegion, defendingRegion); err != nil {
		return fmt.Errorf("region ownership check failed: %w", err)
	}

	if err := checkTroops(ctx, attackingRegion, defendingRegion, move); err != nil {
		return fmt.Errorf("troops check failed: %w", err)
	}

	areNeighbours, err := s.boardService.AreNeighbours(
		ctx,
		attackingRegion.ExternalReference,
		defendingRegion.ExternalReference,
	)
	if err != nil {
		return fmt.Errorf("unable to check if regions are neighbours: %w", err)
	}

	if !areNeighbours {
		return errors.New("attacking region cannot reach defending region")
	}

	ctx.Log().Infow("attack move validation passed", "move", move)

	return nil
}

func checkTroops(
	ctx ctx.MoveContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	ctx.Log().Infow("checking troops")

	if move.AttackingTroops < 1 {
		return errors.New("at least one troop is required to attack")
	}

	if attackingRegion.Troops <= move.AttackingTroops {
		return errors.New("attacking region does not have enough troops")
	}

	if defendingRegion.Troops < 1 {
		ctx.Log().Errorw(
			"attempting to attack a region with no troops",
			"region",
			defendingRegion.ExternalReference,
		)

		return errors.New("defending region does not have enough troops")
	}

	if err := checkDeclaredValues(ctx, attackingRegion, defendingRegion, move); err != nil {
		return fmt.Errorf("declared values are invalid: %w", err)
	}

	ctx.Log().Infow("troops check passed")

	return nil
}

func checkRegionOwnership(
	ctx ctx.MoveContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
) error {
	ctx.Log().Infow("checking region ownership")

	if attackingRegion.UserID != ctx.UserID() {
		return errors.New("attacking region is not owned by player")
	}

	if defendingRegion.UserID == ctx.UserID() {
		return errors.New("cannot attack your own region")
	}

	ctx.Log().Infow("region ownership check passed")

	return nil
}

func checkDeclaredValues(
	ctx ctx.MoveContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	ctx.Log().Infow("checking declared values")

	if attackingRegion.Troops != move.TroopsInSource {
		return errors.New("attacking region doesn't have the declared number of troops")
	}

	if defendingRegion.Troops != move.TroopsInTarget {
		return errors.New("defending region doesn't have the declared number of troops")
	}

	ctx.Log().Infow("declared values check passed")

	return nil
}
