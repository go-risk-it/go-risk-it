package cards

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	targetPhase sqlc.PhaseType,
	_ Move,
) error {
	if targetPhase != sqlc.PhaseTypeDEPLOY {
		return fmt.Errorf("cannot advance cards phase to %s", targetPhase)
	}

	phase, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to insert phase: %w", err)
	}

	deployableTroops := int64(3)

	deployPhase, err := querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          phase.ID,
		DeployableTroops: deployableTroops,
	})
	if err != nil {
		return fmt.Errorf("failed to create deploy phase: %w", err)
	}

	ctx.Log().Infow("deploy phase created", "phase", deployPhase)

	return nil
}
