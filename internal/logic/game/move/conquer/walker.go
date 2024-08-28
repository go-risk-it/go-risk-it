package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) Walk(ctx ctx.MoveContext, querier db.Querier) (sqlc.PhaseType, error) {
	canContinueAttacking, err := s.attackService.CanContinueAttackingQ(ctx, querier)
	if err != nil {
		return "", fmt.Errorf("failed to check if can continue attacking: %w", err)
	}

	if !canContinueAttacking {
		return sqlc.PhaseTypeREINFORCE, nil
	}

	return sqlc.PhaseTypeATTACK, nil
}
