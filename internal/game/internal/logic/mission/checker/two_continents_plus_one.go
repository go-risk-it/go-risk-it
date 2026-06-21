package checker

import (
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
)

type twoContinentsPlusOneChecker struct{}

var _ MissionChecker = (*twoContinentsPlusOneChecker)(nil)

func NewTwoContinentsPlusOneChecker() MissionChecker {
	return &twoContinentsPlusOneChecker{}
}

func (c *twoContinentsPlusOneChecker) Type() snapshot.MissionType {
	return snapshot.MissionTwoContinentsPlusOne
}

func (c *twoContinentsPlusOneChecker) Check(
	checkCtx CheckContext,
	mission snapshot.PlayerMission,
) (bool, error) {
	detail, ok := mission.Detail.(snapshot.TwoContinentsPlusOneMission)
	if !ok {
		return false, fmt.Errorf(
			"expected TwoContinentsPlusOneMission detail, got %T",
			mission.Detail,
		)
	}

	controlled := continentsControlledByPlayer(checkCtx)

	playerControlsTwoMandatoryContinents := slices.ContainsFunc(
		controlled,
		continentEquals(detail.Continent1),
	) &&
		slices.ContainsFunc(controlled, continentEquals(detail.Continent2))

	return playerControlsTwoMandatoryContinents && len(controlled) > 2, nil
}
