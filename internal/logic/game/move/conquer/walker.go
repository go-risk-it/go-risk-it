package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

func (s *service) Walk(
	ctx ctx.GameContext,
	querier db.Querier,
	_ bool,
) (sqlc.GamePhaseType, error) {
	canContinueAttacking, err := s.attackService.CanContinueAttacking(ctx, querier)
	if err != nil {
		return "", fmt.Errorf("failed to check if can continue attacking: %w", err)
	}

	if !canContinueAttacking {
		return sqlc.GamePhaseTypeREINFORCE, nil
	}

	return sqlc.GamePhaseTypeATTACK, nil
}
