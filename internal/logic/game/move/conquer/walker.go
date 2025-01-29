package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

func (s *ServiceImpl) WalkQ(
	ctx ctx.GameContext,
	querier db.Querier,
	_ bool,
) (sqlc.PhaseType, error) {
	canContinueAttacking, err := s.attackService.CanContinueAttackingQ(ctx, querier)
	if err != nil {
		return "", fmt.Errorf("failed to check if can continue attacking: %w", err)
	}

	if !canContinueAttacking {
		return sqlc.PhaseTypeREINFORCE, nil
	}

	return sqlc.PhaseTypeATTACK, nil
}
