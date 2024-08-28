package conquer

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
	_ *MoveResult,
) error {
	if targetPhase != sqlc.PhaseTypeATTACK && targetPhase != sqlc.PhaseTypeREINFORCE {
		return fmt.Errorf("cannot advance conquer phase to %s", targetPhase)
	}

	_, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	return nil
}
