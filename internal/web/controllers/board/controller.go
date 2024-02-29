package board

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/api/game/message"
	sqlc "github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/region"
	"go.uber.org/zap"
)

type Controller interface {
	GetBoardState(ctx context.Context, gameID int64) (message.BoardState, error)
}

type ControllerImpl struct {
	log           *zap.SugaredLogger
	boardService  board.Service
	regionService region.Service
}

func New(
	log *zap.SugaredLogger,
	boardService board.Service,
	regionService region.Service,
) *ControllerImpl {
	return &ControllerImpl{log: log, boardService: boardService, regionService: regionService}
}

func (c *ControllerImpl) GetBoardState(
	ctx context.Context, gameID int64,
) (message.BoardState, error) {
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
