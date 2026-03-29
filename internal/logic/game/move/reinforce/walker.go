package reinforce

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
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
		slog.DebugContext(ctx, "no valid combination, advancing to deploy phase")

		return sqlc.GamePhaseTypeDEPLOY, nil
	}

	slog.DebugContext(ctx, "player has at least one valid combination, advancing to cards phase")

	return sqlc.GamePhaseTypeCARDS, nil
}
