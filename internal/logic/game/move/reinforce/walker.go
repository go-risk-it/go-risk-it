package reinforce

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
) (sqlc.GamePhaseType, error) {
	hasValidCombination, err := s.cardsService.NextPlayerHasValidCombinationQ(ctx, querier)
	if err != nil {
		return "", fmt.Errorf("failed to check if has valid combination: %w", err)
	}

	if !hasValidCombination {
		ctx.Log().Debugw("no valid combination, advancing to deploy phase")

		return sqlc.GamePhaseTypeDEPLOY, nil
	}

	ctx.Log().Debugw("player has at least one valid combination, advancing to cards phase")

	return sqlc.GamePhaseTypeCARDS, nil
}
