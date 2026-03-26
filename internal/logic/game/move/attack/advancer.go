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

func (s *service) Advance(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	performResult *MoveResult,
) error {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeATTACK, targetPhase); err != nil {
		return fmt.Errorf("invalid phase transition: %w", err)
	}

	phase, err := s.phaseService.InsertPhase(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	if targetPhase == sqlc.GamePhaseTypeCONQUER {
		return s.advanceToConquerPhase(ctx, querier, *phase, performResult)
	}

	return nil
}

func (s *service) advanceToConquerPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	phase sqlc.GamePhase,
	performResult *MoveResult,
) error {
	if performResult == nil {
		return errors.New("no attack result available for conquer phase creation")
	}

	if _, err := querier.InsertConquerPhase(ctx, sqlc.InsertConquerPhaseParams{
		PhaseID:             phase.ID,
		ID:                  ctx.GameID(),
		ExternalReference:   performResult.AttackingRegionID,
		ExternalReference_2: performResult.DefendingRegionID,
		MinimumTroops:       performResult.ConqueringTroops,
	}); err != nil {
		return fmt.Errorf("failed to create conquer phase: %w", err)
	}

	slog.DebugContext(ctx, "created conquer phase")

	return nil
}
