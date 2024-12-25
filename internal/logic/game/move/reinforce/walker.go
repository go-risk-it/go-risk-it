package reinforce

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) Walk(ctx ctx.GameContext, querier db.Querier) (sqlc.PhaseType, error) {
	hasValidCombination, err := s.cardsService.HasValidCombination(ctx, querier)
	if err != nil {
		return "", fmt.Errorf("failed to check if has valid combination: %w", err)
	}

	if !hasValidCombination {
		return sqlc.PhaseTypeDEPLOY, nil
	}

	return sqlc.PhaseTypeCARDS, nil
}
