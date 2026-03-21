package attack

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	performResult *MoveResult,
) error {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeATTACK, targetPhase); err != nil {
		return fmt.Errorf("invalid phase transition: %w", err)
	}

	phase, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	if targetPhase == sqlc.GamePhaseTypeCONQUER {
		return s.advanceToConquerPhase(ctx, querier, performResult, *phase)
	}

	return nil
}

func (s *ServiceImpl) advanceToConquerPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	performResult *MoveResult,
	phase sqlc.GamePhase,
) error {
	if _, err := querier.InsertConquerPhase(ctx, sqlc.InsertConquerPhaseParams{
		PhaseID:             phase.ID,
		ID:                  ctx.GameID(),
		ExternalReference:   performResult.AttackingRegionID,
		ExternalReference_2: performResult.DefendingRegionID,
		MinimumTroops:       performResult.ConqueringTroops,
	}); err != nil {
		return fmt.Errorf("failed to create conquer phase: %w", err)
	}

	ctx.Log().Debugw("created conquer phase")

	return nil
}
