package controller

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/api/game/message"
	sqlc "github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/region"
	"go.uber.org/zap"
)

type BoardController interface {
	GetBoardState(ctx context.Context, gameID int64) (message.BoardState, error)
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

func (c *BoardControllerImpl) GetBoardState(
	ctx context.Context, gameID int64,
) (message.BoardState, error) {
	c.log.Infow("getting board state", "gameID", gameID)

	regions, err := c.regionService.GetRegions(ctx, gameID)
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
		OwnerID: region.PlayerName,
		Troops:  region.Troops,
	}
}
