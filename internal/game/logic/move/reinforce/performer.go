package reinforce

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/validation"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// minTroopsToReinforce is the minimum number of troops that must be moved in a reinforcement.
	minTroopsToReinforce = 1
)

func (s *service) Perform(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
) (struct{}, error) {
	observe.SpanEvent(ctx, "performing_reinforce_move",
		attribute.String("source_region_id", move.SourceRegionID),
		attribute.String("target_region_id", move.TargetRegionID),
		attribute.Int64("moving_troops", move.MovingTroops),
	)

	sourceRegion, err := s.regionService.GetRegion(ctx, querier, move.SourceRegionID)
	if err != nil {
		return struct{}{}, fmt.Errorf("unable to get source region: %w", err)
	}

	targetRegion, err := s.regionService.GetRegion(ctx, querier, move.TargetRegionID)
	if err != nil {
		return struct{}{}, fmt.Errorf("unable to get target region: %w", err)
	}

	if err := s.validate(ctx, querier, sourceRegion, targetRegion, move); err != nil {
		return struct{}{}, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.perform(ctx, querier, sourceRegion, targetRegion, move.MovingTroops); err != nil {
		return struct{}{}, fmt.Errorf("unable to perform attack move: %w", err)
	}

	return struct{}{}, nil
}

func (s *service) perform(
	ctx ctx.GameContext,
	querier db.Querier,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
	movingTroops int64,
) error {
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

	return nil
}

func checkRegionOwnership(
	ctx ctx.GameContext,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
) error {
	if err := validation.CheckSourceOwnedByPlayer(ctx, sourceRegion, "source"); err != nil {
		return err
	}

	if err := validation.CheckTargetOwnedByPlayer(ctx, targetRegion); err != nil {
		return err
	}

	return nil
}

func checkTroops(
	ctx ctx.GameContext,
	sourceRegion *sqlc.GetRegionsByGameRow,
	targetRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	if move.MovingTroops < minTroopsToReinforce {
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

	return nil
}
