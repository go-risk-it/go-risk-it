package cards

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.PhaseType,
	moveResult *MoveResult,
) error {
	if targetPhase != sqlc.PhaseTypeDEPLOY {
		return fmt.Errorf("cannot advance cards phase to %s", targetPhase)
	}

	regions, err := s.regionService.GetRegionsQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to get regions: %w", err)
	}

	phase, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	extraDeployableTroops := int64(0)
	if moveResult != nil {
		extraDeployableTroops = moveResult.ExtraDeployableTroops
	}

	// add continents
	if _, err = querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          phase.ID,
		DeployableTroops: int64(countPlayerRegions(ctx, regions)/3) + extraDeployableTroops,
	}); err != nil {
		return fmt.Errorf("failed to create deploy phase: %w", err)
	}

	ctx.Log().Infow("created deploy phase")

	return nil
}

func countPlayerRegions(ctx ctx.GameContext, regions []sqlc.GetRegionsByGameRow) int {
	result := 0

	for _, region := range regions {
		if region.UserID == ctx.UserID() {
			result++
		}
	}

	return result
}
