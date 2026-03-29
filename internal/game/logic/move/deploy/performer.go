package deploy

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/validation"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

const (
	// minTroopsToDeploy is the minimum number of troops that must be deployed in a single move.
	minTroopsToDeploy = 1
)

func (s *service) Perform(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
) (struct{}, error) {
	slog.DebugContext(ctx, "performing deploy move", "move", move)

	deployableTroops, err := s.GetDeployableTroopsWithQuerier(ctx, querier)
	if err != nil {
		return struct{}{}, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	troops := move.DesiredTroops - move.CurrentTroops
	if deployableTroops < troops {
		return struct{}{}, domainerrors.NewValidationError("not enough deployable troops")
	}

	thisRegion, err := s.regionService.GetRegion(ctx, querier, move.RegionID)
	if err != nil {
		return struct{}{}, fmt.Errorf("failed to get region: %w", err)
	}

	if troops < minTroopsToDeploy {
		return struct{}{}, domainerrors.NewValidationError("must deploy at least 1 troop")
	}

	if err := validation.CheckSourceOwnedByPlayer(ctx, thisRegion, "deploy"); err != nil {
		return struct{}{}, err
	}

	if thisRegion.Troops != move.CurrentTroops {
		return struct{}{}, domainerrors.NewValidationError(
			"region has different number of troops than declared",
		)
	}

	if err := s.executeDeploy(ctx, querier, thisRegion, troops); err != nil {
		return struct{}{}, fmt.Errorf("failed to execute deploy: %w", err)
	}

	return struct{}{}, nil
}

func (s *service) executeDeploy(
	ctx ctx.GameContext,
	querier db.Querier,
	region *sqlc.GetRegionsByGameRow,
	troops int64,
) error {
	slog.DebugContext(ctx,
		"executing deploy",
		"region",
		region.ExternalReference,
		"troops",
		troops,
	)

	if err := s.decreaseDeployableTroops(ctx, querier, troops); err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegion(ctx, querier, region, troops); err != nil {
		return fmt.Errorf("failed to increase region troops: %w", err)
	}

	slog.DebugContext(ctx,
		"deploy executed successfully",
		"region",
		region.ExternalReference,
		"troops",
		troops,
	)

	return nil
}

func (s *service) decreaseDeployableTroops(
	ctx ctx.GameContext,
	querier db.Querier,
	troops int64,
) error {
	slog.DebugContext(ctx, "decreasing deployable troops", "troops", troops)

	err := querier.DecreaseDeployableTroops(ctx, sqlc.DecreaseDeployableTroopsParams{
		ID:               ctx.GameID(),
		DeployableTroops: troops,
	})
	if err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	slog.DebugContext(ctx, "decreased deployable troops", "troops", troops)

	return nil
}
