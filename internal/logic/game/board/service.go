package board

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board/dto"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board/graph"
	"go.uber.org/zap"
)

type Service interface {
	GetBoardRegions(ctx ctx.LogContext) ([]string, error)
	AreNeighbours(context ctx.LogContext, source string, target string) bool
	CanPlayerReach(context ctx.MoveContext, source string, target string) bool
}

type ServiceImpl struct {
	log   *zap.SugaredLogger
	graph graph.Graph
}

var _ Service = (*ServiceImpl)(nil)

func (s *ServiceImpl) AreNeighbours(context ctx.LogContext, source string, target string) bool {
	graph, err := s.getGraph(context)
	if err != nil {
		return false
	}

	return graph.AreNeighbours(source, target)
}

func (s *ServiceImpl) CanPlayerReach(context ctx.MoveContext, source string, target string) bool {
	return false
}

var _ Service = (*ServiceImpl)(nil)

func NewService(logger *zap.SugaredLogger) *ServiceImpl {
	return &ServiceImpl{log: logger, graph: nil}
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

	s.graph = graph.New(boardDto)

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
