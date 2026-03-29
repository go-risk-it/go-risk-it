package checker

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/region"
)

type EighteenTerritoriesChecker struct {
	regionService region.Service
}

var _ MissionChecker = (*EighteenTerritoriesChecker)(nil)

func NewEighteenTerritoriesChecker(regionService region.Service) *EighteenTerritoriesChecker {
	return &EighteenTerritoriesChecker{regionService: regionService}
}

func (c *EighteenTerritoriesChecker) Type() sqlc.GameMissionType {
	return sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS
}

func (c *EighteenTerritoriesChecker) Check(
	ctx ctx.GameContext,
	querier db.Querier,
	_ sqlc.GameMission,
) (bool, error) {
	regions, err := c.regionService.GetPlayerRegions(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get player regions: %w", err)
	}

	count := 0

	for _, r := range regions {
		if r.Troops > 1 {
			count++
		}
	}

	return count >= 18, nil
}
