package controller

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/region"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

type BoardController struct {
	regionService region.Service
}

func NewBoardController(regionService region.Service) *BoardController {
	return &BoardController{regionService: regionService}
}

func (c *BoardController) GetBoardState(ctx ctx.GameContext) (messaging.BoardState, error) {
	slog.InfoContext(ctx, "getting board state")

	regions, err := c.regionService.GetRegions(ctx)
	if err != nil {
		return messaging.BoardState{}, fmt.Errorf("unable to get regions: %w", err)
	}

	return messaging.BoardState{Regions: convertRegions(regions)}, nil
}

func convertRegions(regions []sqlc.GetRegionsByGameRow) []messaging.Region {
	result := make([]messaging.Region, len(regions))
	for i, r := range regions {
		result[i] = convertRegion(r)
	}

	return result
}

func convertRegion(region sqlc.GetRegionsByGameRow) messaging.Region {
	return messaging.Region{
		ID:      region.ExternalReference,
		OwnerID: region.UserID,
		Troops:  region.Troops,
	}
}
