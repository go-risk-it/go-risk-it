package checker

import (
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
)

type twoContinentsChecker struct{}

var _ MissionChecker = (*twoContinentsChecker)(nil)

func NewTwoContinentsChecker() MissionChecker {
	return &twoContinentsChecker{}
}

func (c *twoContinentsChecker) Type() snapshot.MissionType {
	return snapshot.MissionTwoContinents
}

func (c *twoContinentsChecker) Check(
	checkCtx CheckContext,
	mission snapshot.PlayerMission,
) (bool, error) {
	detail, ok := mission.Detail.(snapshot.TwoContinentsMission)
	if !ok {
		return false, fmt.Errorf(
			"expected TwoContinentsMission detail, got %T",
			mission.Detail,
		)
	}

	controlled := continentsControlledByPlayer(checkCtx)

	return slices.ContainsFunc(controlled, continentEquals(detail.Continent1)) &&
		slices.ContainsFunc(controlled, continentEquals(detail.Continent2)), nil
}

func continentEquals(cont string) func(continent *board.Continent) bool {
	return func(continent *board.Continent) bool {
		return continent.ExternalReference == cont
	}
}

func continentsControlledByPlayer(checkCtx CheckContext) []*board.Continent {
	playerRegions := make([]string, 0)

	for _, r := range checkCtx.Regions {
		if r.OwnerID == checkCtx.CurrentUserID {
			playerRegions = append(playerRegions, r.ID)
		}
	}

	return checkCtx.Continents.GetContinentsControlledBy(playerRegions)
}
