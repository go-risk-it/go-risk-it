package deploy

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
	_ *MoveResult,
) error {
	if targetPhase != sqlc.PhaseTypeATTACK {
		return fmt.Errorf("cannot advance deploy phase to %s", targetPhase)
	}

	_, err := s.phaseService.InsertPhaseQ(ctx, querier, sqlc.PhaseTypeATTACK)
	if err != nil {
		return fmt.Errorf("failed to create attack phase: %w", err)
	}

	return nil
}
