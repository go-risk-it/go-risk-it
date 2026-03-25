package reinforce

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/validation"
)

func (s *service) Perform(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
) (any, error) {
	slog.InfoContext(ctx, "performing reinforce move", "move", move)

	sourceRegion, err := s.regionService.GetRegion(ctx, querier, move.SourceRegionID)
	if err != nil {
		return nil, fmt.Errorf("unable to get source region: %w", err)
	}

	targetRegion, err := s.regionService.GetRegion(ctx, querier, move.TargetRegionID)
	if err != nil {
		return nil, fmt.Errorf("unable to get target region: %w", err)
	}

	if err := s.validate(ctx, querier, sourceRegion, targetRegion, move); err != nil {
		slog.InfoContext(ctx, "validation failed", "error", err)

		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.perform(ctx, querier, sourceRegion, targetRegion, move.MovingTroops); err != nil {
		return nil, fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil, nil //nolint:nilnil // no result needed for reinforce
}

func (s *service) perform(
	ctx ctx.GameContext,
	querier db.Querier,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
	movingTroops int64,
) error {
	slog.InfoContext(ctx, "updating region troops")

	if err := s.regionService.UpdateTroopsInRegion(
		ctx,
		querier,
		sourceRegion,
		-movingTroops,
	); err != nil {
		return fmt.Errorf("failed to decrease troops in attacking region: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegion(
		ctx,
		querier,
		targetRegion,
		movingTroops,
	); err != nil {
		return fmt.Errorf("failed to decrease troops in defending region: %w", err)
	}

	return nil
}

func (s *service) validate(
	ctx ctx.GameContext,
	querier db.Querier,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	slog.InfoContext(ctx, "validating reinforce move")

	if err := checkRegionOwnership(ctx, sourceRegion, targetRegion); err != nil {
		return fmt.Errorf("region ownership check failed: %w", err)
	}

	if err := checkTroops(ctx, sourceRegion, targetRegion, move); err != nil {
		return fmt.Errorf("troops check failed: %w", err)
	}

	canReach, err := s.boardService.CanPlayerReach(
		ctx,
		querier,
		sourceRegion.ExternalReference,
		targetRegion.ExternalReference,
	)
	if err != nil {
		return fmt.Errorf("failed to check if player can reach target: %w", err)
	}

	if !canReach {
		return domainerrors.NewValidationError("player cannot reach target region")
	}

	slog.InfoContext(ctx, "reinforce move validation passed")

	return nil
}

func checkRegionOwnership(
	ctx ctx.GameContext,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
) error {
	slog.InfoContext(ctx, "checking region ownership")

	if sourceRegion.UserID != ctx.UserID() {
		return domainerrors.NewValidationError("source region is not owned by player")
	}

	if targetRegion.UserID != ctx.UserID() {
		return domainerrors.NewValidationError("target region is not owned by player")
	}

	slog.InfoContext(ctx, "region ownership check passed")

	return nil
}

func checkTroops(
	ctx ctx.GameContext,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	slog.InfoContext(ctx, "checking troops")

	if move.MovingTroops < 1 {
		return domainerrors.NewValidationError("at least one troop is required to reinforce")
	}

	if sourceRegion.Troops <= move.MovingTroops {
		return domainerrors.NewValidationError("source region does not have enough troops")
	}

	if err := validation.CheckDeclaredTroops(
		ctx,
		sourceRegion.Troops,
		targetRegion.Troops,
		move.TroopsInSource,
		move.TroopsInTarget,
	); err != nil {
		return fmt.Errorf("declared values are invalid: %w", err)
	}

	slog.InfoContext(ctx, "troops check passed")

	return nil
}
