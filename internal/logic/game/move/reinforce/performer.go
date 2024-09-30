package reinforce

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
	ctx.Log().Infow("performing reinforce move", "move", move)

	sourceRegion, err := s.regionService.GetRegionQ(ctx, querier, move.SourceRegionID)
	if err != nil {
		return nil, fmt.Errorf("unable to get source region: %w", err)
	}

	targetRegion, err := s.regionService.GetRegionQ(ctx, querier, move.TargetRegionID)
	if err != nil {
		return nil, fmt.Errorf("unable to get target region: %w", err)
	}

	if err := s.validate(ctx, querier, sourceRegion, targetRegion, move); err != nil {
		ctx.Log().Infow("validation failed", "error", err)

		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.perform(ctx, querier, sourceRegion, targetRegion, move.MovingTroops); err != nil {
		return nil, fmt.Errorf("unable to perform attack move: %w", err)
	}

	return &MoveResult{}, nil
}

func (s *ServiceImpl) perform(
	ctx ctx.GameContext,
	querier db.Querier,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
	movingTroops int64,
) error {
	ctx.Log().Infow("updating region troops")

	if err := s.regionService.UpdateTroopsInRegionQ(
		ctx,
		querier,
		sourceRegion,
		-movingTroops,
	); err != nil {
		return fmt.Errorf("failed to decrease troops in attacking region: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegionQ(
		ctx,
		querier,
		targetRegion,
		movingTroops,
	); err != nil {
		return fmt.Errorf("failed to decrease troops in defending region: %w", err)
	}

	return nil
}

func (s *ServiceImpl) validate(
	ctx ctx.GameContext,
	querier db.Querier,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	ctx.Log().Infow("validating reinforce move")

	if err := checkRegionOwnership(ctx, sourceRegion, targetRegion); err != nil {
		return fmt.Errorf("region ownership check failed: %w", err)
	}

	if err := checkTroops(ctx, sourceRegion, targetRegion, move); err != nil {
		return fmt.Errorf("troops check failed: %w", err)
	}

	canReach, err := s.boardService.CanPlayerReachQ(
		ctx,
		querier,
		sourceRegion.ExternalReference,
		targetRegion.ExternalReference,
	)
	if err != nil {
		return fmt.Errorf("failed to check if player can reach target: %w", err)
	}

	if !canReach {
		return errors.New("player cannot reach target region")
	}

	ctx.Log().Infow("reinforce move validation passed")

	return nil
}

func checkRegionOwnership(
	ctx ctx.GameContext,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
) error {
	ctx.Log().Infow("checking region ownership")

	if sourceRegion.UserID != ctx.UserID() {
		return errors.New("source region is not owned by player")
	}

	if targetRegion.UserID != ctx.UserID() {
		return errors.New("target region is not owned by player")
	}

	ctx.Log().Infow("region ownership check passed")

	return nil
}

func checkTroops(
	ctx ctx.GameContext,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	ctx.Log().Infow("checking troops")

	if move.MovingTroops < 1 {
		return errors.New("at least one troop is required to reinforce")
	}

	if sourceRegion.Troops <= move.MovingTroops {
		return errors.New("source region does not have enough troops")
	}

	if err := checkDeclaredValues(ctx, sourceRegion, targetRegion, move); err != nil {
		return fmt.Errorf("declared values are invalid: %w", err)
	}

	ctx.Log().Infow("troops check passed")

	return nil
}

func checkDeclaredValues(
	ctx ctx.GameContext,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	ctx.Log().Infow("checking declared values")

	if sourceRegion.Troops != move.TroopsInSource {
		return errors.New("source region doesn't have the declared number of troops")
	}

	if targetRegion.Troops != move.TroopsInTarget {
		return errors.New("target region doesn't have the declared number of troops")
	}

	ctx.Log().Infow("declared values check passed")

	return nil
}
