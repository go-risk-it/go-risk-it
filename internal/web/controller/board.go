package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/message"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"go.uber.org/zap"
)

type BoardController interface {
	GetBoardState(ctx ctx.GameContext) (message.BoardState, error)
}

type BoardControllerImpl struct {
	log           *zap.SugaredLogger
	boardService  board.Service
	regionService region.Service
}

func NewBoardController(
	log *zap.SugaredLogger,
	boardService board.Service,
	regionService region.Service,
) *BoardControllerImpl {
	return &BoardControllerImpl{log: log, boardService: boardService, regionService: regionService}
}

func (c *BoardControllerImpl) GetBoardState(ctx ctx.GameContext) (message.BoardState, error) {
	c.log.Infow("getting board state")

	regions, err := c.regionService.GetRegions(ctx)
	if err != nil {
		return message.BoardState{}, fmt.Errorf("unable to get regions: %w", err)
	}

	return message.BoardState{Regions: convertRegions(regions)}, nil
}

func convertRegions(regions []sqlc.GetRegionsByGameRow) []message.Region {
	result := make([]message.Region, len(regions))
	for i, r := range regions {
		result[i] = convertRegion(r)
	}

	return result
}

func convertRegion(region sqlc.GetRegionsByGameRow) message.Region {
	return message.Region{
		ID:      region.ExternalReference,
		OwnerID: region.UserID,
		Troops:  region.Troops,
	}
}
