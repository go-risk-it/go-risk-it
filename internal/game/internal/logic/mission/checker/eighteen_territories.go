package checker

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
)

type eighteenTerritoriesChecker struct{}

var _ MissionChecker = (*eighteenTerritoriesChecker)(nil)

func NewEighteenTerritoriesChecker() MissionChecker {
	return &eighteenTerritoriesChecker{}
}

func (c *eighteenTerritoriesChecker) Type() snapshot.MissionType {
	return snapshot.MissionEighteenTerritoriesTwoTroops
}

func (c *eighteenTerritoriesChecker) Check(
	checkCtx CheckContext,
	_ snapshot.PlayerMission,
) (bool, error) {
	count := 0

	for _, r := range checkCtx.Regions {
		if r.OwnerID == checkCtx.CurrentUserID && r.Troops > 1 {
			count++
		}
	}

	return count >= 18, nil
}
