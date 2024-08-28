package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
)

func (s *ServiceImpl) PerformQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	move Move,
) (*MoveResult, error) {
	ctx.Log().Infow("performing conquer move", "move", move)

	phaseState, err := s.GetPhaseStateQ(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("unable to get phase state: %w", err)
	}

	if phaseState.MinimumTroops > move.Troops {
		return nil, fmt.Errorf("must move at least %d troops", phaseState.MinimumTroops)
	}

	sourceRegion, err := s.regionService.GetRegionQ(ctx, querier, phaseState.SourceRegion)
	if err != nil {
		return nil, fmt.Errorf("unable to get attacking region: %w", err)
	}

	targetRegion, err := s.regionService.GetRegionQ(ctx, querier, phaseState.TargetRegion)
	if err != nil {
		return nil, fmt.Errorf("unable to get defending region: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegion(
		ctx,
		querier,
		sourceRegion,
		-move.Troops,
	); err != nil {
		return nil, fmt.Errorf("failed to decrease troops in source region: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegion(
		ctx,
		querier,
		targetRegion,
		move.Troops,
	); err != nil {
		return nil, fmt.Errorf("failed to increase troops in target region: %w", err)
	}

	ctx.Log().Infow("conquer executed successfully")

	return &MoveResult{}, nil
}
