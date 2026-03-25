package deploy

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
)

func (s *service) Perform(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
) (any, error) {
	slog.InfoContext(ctx, "performing deploy move", "move", move)

	deployableTroops, err := s.GetDeployableTroopsWithQuerier(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	troops := move.DesiredTroops - move.CurrentTroops
	if deployableTroops < troops {
		return nil, domainerrors.NewValidationError("not enough deployable troops")
	}

	thisRegion, err := s.regionService.GetRegion(ctx, querier, move.RegionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	if troops < 1 {
		return nil, domainerrors.NewValidationError("must deploy at least 1 troop")
	}

	if thisRegion.UserID != ctx.UserID() {
		return nil, domainerrors.NewValidationError("region is not owned by player")
	}

	if thisRegion.Troops != move.CurrentTroops {
		return nil, domainerrors.NewValidationError(
			"region has different number of troops than declared",
		)
	}

	if err := s.executeDeploy(ctx, querier, thisRegion, troops); err != nil {
		return nil, fmt.Errorf("failed to execute deploy: %w", err)
	}

	return nil, nil //nolint:nilnil // no result needed for deploy
}

func (s *service) executeDeploy(
	ctx ctx.GameContext,
	querier db.Querier,
	region *sqlc.GetRegionsByGameRow,
	troops int64,
) error {
	slog.InfoContext(ctx,
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

	slog.InfoContext(ctx,
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
	slog.InfoContext(ctx, "decreasing deployable troops", "troops", troops)

	err := querier.DecreaseDeployableTroops(ctx, sqlc.DecreaseDeployableTroopsParams{
		ID:               ctx.GameID(),
		DeployableTroops: troops,
	})
	if err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	slog.InfoContext(ctx, "decreased deployable troops", "troops", troops)

	return nil
}
