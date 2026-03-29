package deploy

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
	deployableTroops, err := s.GetDeployableTroopsWithQuerier(ctx, querier)
	if err != nil {
		return sqlc.GamePhaseTypeDEPLOY, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	if deployableTroops == 0 {
		return sqlc.GamePhaseTypeATTACK, nil
	}

	return sqlc.GamePhaseTypeDEPLOY, nil
}
