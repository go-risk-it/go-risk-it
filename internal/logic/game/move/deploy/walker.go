package deploy

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
	deployableTroops, err := s.GetDeployableTroopsQ(ctx, querier)
	if err != nil {
		return sqlc.GamePhaseTypeDEPLOY, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	if deployableTroops == 0 {
		return sqlc.GamePhaseTypeATTACK, nil
	}

	return sqlc.GamePhaseTypeDEPLOY, nil
}
