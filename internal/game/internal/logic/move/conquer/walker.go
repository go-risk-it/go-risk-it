package conquer

import (
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/attack"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

func (s *service) Walk(wctx moveservice.WalkContext) (sqlc.GamePhaseType, error) {
	regions := attack.MergeRegions(wctx.PrevSnapshot.Regions, wctx.Effect.RegionUpdates)

	if !attack.CanContinueAttacking(regions, wctx.CurrentUserID) {
		return sqlc.GamePhaseTypeREINFORCE, nil
	}

	return sqlc.GamePhaseTypeATTACK, nil
}
