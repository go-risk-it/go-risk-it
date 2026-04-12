package deploy

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

func (s *service) Walk(wctx moveservice.WalkContext) (sqlc.GamePhaseType, error) {
	deployState, ok := wctx.Effect.UpdatedPhase.(snapshot.DeployPhaseState)
	if !ok {
		return "", fmt.Errorf("expected DeployPhaseState, got %T", wctx.Effect.UpdatedPhase)
	}

	if deployState.DeployableTroops <= 0 {
		return sqlc.GamePhaseTypeATTACK, nil
	}

	return sqlc.GamePhaseTypeDEPLOY, nil
}
