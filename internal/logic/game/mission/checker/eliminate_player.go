package checker

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
)

type EliminatePlayerChecker struct {
	regionService region.Service
}

var _ MissionChecker = (*EliminatePlayerChecker)(nil)

func NewEliminatePlayerChecker(regionService region.Service) *EliminatePlayerChecker {
	return &EliminatePlayerChecker{regionService: regionService}
}

func (c *EliminatePlayerChecker) Type() sqlc.GameMissionType {
	return sqlc.GameMissionTypeELIMINATEPLAYER
}

func (c *EliminatePlayerChecker) Check(
	ctx ctx.GameContext,
	querier db.Querier,
	baseMission sqlc.GameMission,
) (bool, error) {
	mission, err := querier.GetEliminatePlayerMission(ctx, baseMission.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get eliminate player mission: %w", err)
	}

	targetPlayerRegions, err := c.regionService.GetRegionsControlledByPlayer(
		ctx,
		querier,
		mission.TargetPlayerID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to get player regions: %w", err)
	}

	return len(targetPlayerRegions) == 0, nil
}
