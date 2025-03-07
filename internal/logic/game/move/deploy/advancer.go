package deploy

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
	if targetPhase != sqlc.GamePhaseTypeATTACK {
		return fmt.Errorf("cannot advance deploy phase to %s", targetPhase)
	}

	_, err := s.phaseService.InsertPhaseQ(ctx, querier, sqlc.GamePhaseTypeATTACK)
	if err != nil {
		return fmt.Errorf("failed to create attack phase: %w", err)
	}

	return nil
}
