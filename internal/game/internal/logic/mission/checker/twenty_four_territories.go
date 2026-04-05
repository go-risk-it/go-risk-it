package checker

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/region"
)

type TwentyFourTerritoriesChecker struct {
	regionService region.Service
}

var _ MissionChecker = (*TwentyFourTerritoriesChecker)(nil)

func NewTwentyFourTerritoriesChecker(regionService region.Service) *TwentyFourTerritoriesChecker {
	return &TwentyFourTerritoriesChecker{regionService: regionService}
}

func (c *TwentyFourTerritoriesChecker) Type() sqlc.GameMissionType {
	return sqlc.GameMissionTypeTWENTYFOURTERRITORIES
}

func (c *TwentyFourTerritoriesChecker) Check(
	ctx ctx.GameContext,
	querier db.Querier,
	_ sqlc.GameMission,
) (bool, error) {
	regions, err := c.regionService.GetPlayerRegions(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get player regions: %w", err)
	}

	return len(regions) >= 24, nil
}
