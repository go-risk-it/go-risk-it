package attack

import (
	"errors"
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
) error {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeATTACK, targetPhase); err != nil {
		return fmt.Errorf("invalid phase transition: %w", err)
	}

	phase, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	if targetPhase == sqlc.GamePhaseTypeCONQUER {
		return s.advanceToConquerPhase(ctx, querier, *phase)
	}

	return nil
}

func (s *ServiceImpl) advanceToConquerPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	phase sqlc.GamePhase,
) error {
	if s.lastResult == nil {
		return errors.New("no attack result available for conquer phase creation")
	}

	if _, err := querier.InsertConquerPhase(ctx, sqlc.InsertConquerPhaseParams{
		PhaseID:             phase.ID,
		ID:                  ctx.GameID(),
		ExternalReference:   s.lastResult.AttackingRegionID,
		ExternalReference_2: s.lastResult.DefendingRegionID,
		MinimumTroops:       s.lastResult.ConqueringTroops,
	}); err != nil {
		return fmt.Errorf("failed to create conquer phase: %w", err)
	}

	ctx.Log().Debugw("created conquer phase")

	return nil
}
