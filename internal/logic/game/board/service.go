package board

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board/dto"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board/graph"
	"go.uber.org/zap"
)

type Service interface {
	GetBoardRegions(ctx ctx.LogContext) ([]string, error)
	AreNeighbours(ctx ctx.LogContext, source string, target string) (bool, error)
	CanPlayerReachQ(
		ctx ctx.GameContext,
		querier db.Querier,
		source string,
		target string,
	) (bool, error)
}

type ServiceImpl struct {
	log           *zap.SugaredLogger
	graph         graph.Graph
	regionService region.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(logger *zap.SugaredLogger, regionService region.Service) *ServiceImpl {
	return &ServiceImpl{log: logger, graph: nil, regionService: regionService}
}

func (s *ServiceImpl) AreNeighbours(
	ctx ctx.LogContext,
	source string,
	target string,
) (bool, error) {
	graph, err := s.getGraph(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get graph: %w", err)
	}

	return graph.AreNeighbours(source, target), nil
}

func (s *ServiceImpl) CanPlayerReachQ(
	ctx ctx.GameContext,
	querier db.Querier,
	source string,
	target string,
) (bool, error) {
	ctx.Log().Infow("checking if player can reach target", "source", source, "target", target)

	regions, err := s.regionService.GetRegionsQ(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get regions: %w", err)
	}

	usableRegions := make(map[string]struct{})

	for _, region := range regions {
		if region.UserID == ctx.UserID() {
			usableRegions[region.ExternalReference] = struct{}{}
		}
	}

	graph, err := s.getGraph(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get graph: %w", err)
	}

	return graph.CanReach(ctx, source, target, usableRegions), nil
}

func (s *ServiceImpl) GetBoardRegions(ctx ctx.LogContext) ([]string, error) {
	ctx.Log().Infow("getting board regions")

	graph, err := s.getGraph(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	result := graph.GetRegions()

	ctx.Log().Infow("got board regions", "regions", result)

	return result, nil
}

func (s *ServiceImpl) getGraph(ctx ctx.LogContext) (graph.Graph, error) {
	ctx.Log().Infow("getting graph")

	if s.graph != nil {
		ctx.Log().Infow("graph cache hit")

		return s.graph, nil
	}

	ctx.Log().Infow("graph cache miss, fetching board from file")

	boardDto, err := s.fetchFromFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get boardDto: %w", err)
	}

	s.graph, err = graph.New(boardDto)
	if err != nil {
		return nil, fmt.Errorf("failed to create graph: %w", err)
	}

	ctx.Log().Infow("graph cache updated")

	return s.graph, nil
}

func (s *ServiceImpl) fetchFromFile(ctx ctx.LogContext) (*dto.Board, error) {
	data, err := os.ReadFile("map.json")
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	board := &dto.Board{}

	err = json.Unmarshal(data, board)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	ctx.Log().Debugw("Read board from file", "board", board)

	return board, nil
}
