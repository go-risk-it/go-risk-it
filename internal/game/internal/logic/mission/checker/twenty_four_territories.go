package checker

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
)

type twentyFourTerritoriesChecker struct{}

var _ MissionChecker = (*twentyFourTerritoriesChecker)(nil)

func NewTwentyFourTerritoriesChecker() MissionChecker {
	return &twentyFourTerritoriesChecker{}
}

func (c *twentyFourTerritoriesChecker) Type() snapshot.MissionType {
	return snapshot.MissionTwentyFourTerritories
}

func (c *twentyFourTerritoriesChecker) Check(
	checkCtx CheckContext,
	_ snapshot.PlayerMission,
) (bool, error) {
	count := 0

	for _, r := range checkCtx.Regions {
		if r.OwnerID == checkCtx.CurrentUserID {
			count++
		}
	}

	return count >= 24, nil
}
