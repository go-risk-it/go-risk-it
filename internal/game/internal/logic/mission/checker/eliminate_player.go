package checker

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
)

type eliminatePlayerChecker struct{}

var _ MissionChecker = (*eliminatePlayerChecker)(nil)

func NewEliminatePlayerChecker() MissionChecker {
	return &eliminatePlayerChecker{}
}

func (c *eliminatePlayerChecker) Type() snapshot.MissionType {
	return snapshot.MissionEliminatePlayer
}

func (c *eliminatePlayerChecker) Check(
	checkCtx CheckContext,
	mission snapshot.PlayerMission,
) (bool, error) {
	detail, ok := mission.Detail.(snapshot.EliminatePlayerMission)
	if !ok {
		return false, fmt.Errorf(
			"expected EliminatePlayerMission detail, got %T",
			mission.Detail,
		)
	}

	for _, r := range checkCtx.Regions {
		if r.OwnerID == detail.TargetUserID {
			return false, nil
		}
	}

	return true, nil
}
