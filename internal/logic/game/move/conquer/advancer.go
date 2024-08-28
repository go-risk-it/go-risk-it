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
	performResult *MoveResult,
) error {
	if targetPhase != sqlc.PhaseTypeATTACK && targetPhase != sqlc.PhaseTypeREINFORCE {
		return fmt.Errorf("cannot advance conquer phase to %s", targetPhase)
	}

	phase, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	if targetPhase == sqlc.PhaseTypeATTACK {
		return s.advanceToAttackPhase(ctx, querier, performResult, *phase)
	}

	if targetPhase == sqlc.PhaseTypeREINFORCE {
		return s.advanceToReinforcePhase(ctx, querier, performResult, *phase)
	}

	return fmt.Errorf("cannot advance conquer phase to %s", targetPhase)
}

func (s *ServiceImpl) advanceToAttackPhase(
	ctx ctx.MoveContext,
	querier db.Querier,
	performResult *MoveResult,
	phase sqlc.Phase,
) error {
	return nil
}

func (s *ServiceImpl) advanceToReinforcePhase(
	ctx ctx.MoveContext,
	querier db.Querier,
	performResult *MoveResult,
	phase sqlc.Phase,
) error {
	return nil
}
