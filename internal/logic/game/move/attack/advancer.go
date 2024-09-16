package attack

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
	performResult *MoveResult,
) error {
	if targetPhase != sqlc.PhaseTypeCONQUER && targetPhase != sqlc.PhaseTypeREINFORCE {
		return fmt.Errorf("cannot advance attack phase to %s", targetPhase)
	}

	phase, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	if targetPhase == sqlc.PhaseTypeCONQUER {
		return s.advanceToConquerPhase(ctx, querier, performResult, *phase)
	}

	if targetPhase == sqlc.PhaseTypeREINFORCE {
		return s.advanceToReinforcePhase(ctx, querier, performResult, *phase)
	}

	return fmt.Errorf("cannot advance attack phase to %s", targetPhase)
}

func (s *ServiceImpl) advanceToConquerPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	performResult *MoveResult,
	phase sqlc.Phase,
) error {
	conquerPhase, err := querier.InsertConquerPhase(ctx, sqlc.InsertConquerPhaseParams{
		PhaseID:             phase.ID,
		ID:                  ctx.GameID(),
		ExternalReference:   performResult.AttackingRegionID,
		ExternalReference_2: performResult.DefendingRegionID,
		MinimumTroops:       performResult.ConqueringTroops,
	})
	if err != nil {
		return fmt.Errorf("failed to create conquer phase: %w", err)
	}

	ctx.Log().Infow("created conquer phase", "phase", conquerPhase)

	return nil
}

func (s *ServiceImpl) advanceToReinforcePhase(
	ctx ctx.GameContext,
	querier db.Querier,
	performResult *MoveResult,
	phase sqlc.Phase,
) error {
	return nil
}
