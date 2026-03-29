package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
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
