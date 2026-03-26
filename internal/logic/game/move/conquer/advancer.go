package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
)

func (s *service) Advance(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	_ struct{},
) error {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeCONQUER, targetPhase); err != nil {
		return fmt.Errorf("invalid phase transition: %w", err)
	}

	_, err := s.phaseService.InsertPhase(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to create phase: %w", err)
	}

	return nil
}
