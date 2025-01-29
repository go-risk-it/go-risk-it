package deploy

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

func (s *ServiceImpl) PerformQ(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
) (*MoveResult, error) {
	ctx.Log().Infow("performing deploy move", "move", move)

	deployableTroops, err := s.GetDeployableTroopsQ(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	troops := move.DesiredTroops - move.CurrentTroops
	if deployableTroops < troops {
		return nil, errors.New("not enough deployable troops")
	}

	thisRegion, err := s.regionService.GetRegionQ(ctx, querier, move.RegionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	if troops < 1 {
		return nil, errors.New("must deploy at least 1 troop")
	}

	if thisRegion.UserID != ctx.UserID() {
		return nil, errors.New("region is not owned by player")
	}

	if thisRegion.Troops != move.CurrentTroops {
		return nil, errors.New("region has different number of troops than declared")
	}

	if err := s.executeDeploy(ctx, querier, thisRegion, troops); err != nil {
		return nil, fmt.Errorf("failed to execute deploy: %w", err)
	}

	return &MoveResult{}, nil
}

func (s *ServiceImpl) executeDeploy(
	ctx ctx.GameContext,
	querier db.Querier,
	region *sqlc.GetRegionsByGameRow,
	troops int64,
) error {
	ctx.Log().Infow(
		"executing deploy",
		"region",
		region.ExternalReference,
		"troops",
		troops,
	)

	if err := s.decreaseDeployableTroopsQ(ctx, querier, troops); err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegionQ(ctx, querier, region, troops); err != nil {
		return fmt.Errorf("failed to increase region troops: %w", err)
	}

	ctx.Log().Infow(
		"deploy executed successfully",
		"region",
		region.ExternalReference,
		"troops",
		troops,
	)

	return nil
}

func (s *ServiceImpl) decreaseDeployableTroopsQ(
	ctx ctx.GameContext,
	querier db.Querier,
	troops int64,
) error {
	ctx.Log().Infow("decreasing deployable troops", "troops", troops)

	err := querier.DecreaseDeployableTroops(ctx, sqlc.DecreaseDeployableTroopsParams{
		ID:               ctx.GameID(),
		DeployableTroops: troops,
	})
	if err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	ctx.Log().Infow("decreased deployable troops", "troops", troops)

	return nil
}
