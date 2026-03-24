package attack

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	performResult any,
) error {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeATTACK, targetPhase); err != nil {
		return fmt.Errorf("invalid phase transition: %w", err)
	}

	phase, err := s.phaseService.InsertPhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	if targetPhase == sqlc.GamePhaseTypeCONQUER {
		return s.advanceToConquerPhase(ctx, querier, *phase, performResult)
	}

	return nil
}

func (s *ServiceImpl) advanceToConquerPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	phase sqlc.GamePhase,
	performResult any,
) error {
	result, ok := performResult.(*MoveResult)
	if !ok || result == nil {
		return errors.New("no attack result available for conquer phase creation")
	}

	if _, err := querier.InsertConquerPhase(ctx, sqlc.InsertConquerPhaseParams{
		PhaseID:             phase.ID,
		ID:                  ctx.GameID(),
		ExternalReference:   result.AttackingRegionID,
		ExternalReference_2: result.DefendingRegionID,
		MinimumTroops:       result.ConqueringTroops,
	}); err != nil {
		return fmt.Errorf("failed to create conquer phase: %w", err)
	}

	slog.DebugContext(ctx, "created conquer phase")

	return nil
}
