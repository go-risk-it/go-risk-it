package deploy

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

func (s *service) Advance(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	_ struct{},
) error {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeDEPLOY, targetPhase); err != nil {
		return fmt.Errorf("invalid phase transition: %w", err)
	}

	_, err := s.phaseService.InsertPhase(ctx, querier, sqlc.GamePhaseTypeATTACK)
	if err != nil {
		return fmt.Errorf("failed to create attack phase: %w", err)
	}

	return nil
}
