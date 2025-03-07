package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	_ *MoveResult,
) error {
	if targetPhase != sqlc.GamePhaseTypeATTACK && targetPhase != sqlc.GamePhaseTypeREINFORCE {
		return fmt.Errorf("cannot advance conquer phase to %s", targetPhase)
	}

	_, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	return nil
}
