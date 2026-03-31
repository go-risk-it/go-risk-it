package reinforce

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
	hasValidCombination, err := s.cardsService.NextPlayerHasValidCombination(ctx, querier)
	if err != nil {
		return "", fmt.Errorf("failed to check if has valid combination: %w", err)
	}

	if !hasValidCombination {
		return sqlc.GamePhaseTypeDEPLOY, nil
	}

	return sqlc.GamePhaseTypeCARDS, nil
}
