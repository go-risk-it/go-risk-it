package board

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"go.uber.org/zap"
)

type Service interface {
	GetBoardRegions(ctx ctx.LogContext) ([]string, error)
	// AreNeighbours(context ctx.LogContext, source string, target string) bool
	// CanPlayerReach(context ctx.MoveContext, source string, target string) bool
}

type ServiceImpl struct {
	log        *zap.SugaredLogger
	boardCache *BoardDto
}

var _ Service = (*ServiceImpl)(nil)

func NewService(logger *zap.SugaredLogger) *ServiceImpl {
	return &ServiceImpl{log: logger, boardCache: nil}
}

func (s *ServiceImpl) GetBoardRegions(ctx ctx.LogContext) ([]string, error) {
	ctx.Log().Infow("getting board regions")

	board, err := s.getBoardDto(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	ctx.Log().Infow("got board regions", "regions", len(board.Regions))

	result := make([]string, 0, len(board.Regions))
	for _, region := range board.Regions {
		result = append(result, region.ExternalReference)
	}

	return result, nil
}

func (s *ServiceImpl) getBoardDto(ctx ctx.LogContext) (*BoardDto, error) {
	ctx.Log().Infow("getting board")

	if s.boardCache != nil {
		ctx.Log().Infow("board cache hit")

		return s.boardCache, nil
	}

	ctx.Log().Infow("board cache miss, fetching from file")

	boardDto, err := s.fetchFromFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get boardDto: %w", err)
	}

	s.boardCache = boardDto

	ctx.Log().Infow("board cache updated")

	return boardDto, nil
}

func (s *ServiceImpl) fetchFromFile(ctx ctx.LogContext) (*BoardDto, error) {
	data, err := os.ReadFile("map.json")
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	board := &BoardDto{}

	err = json.Unmarshal(data, board)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	ctx.Log().Debugw("Read board from file", "board", board)

	return board, nil
}

// func (s *ServiceImpl) AreNeighbours(context ctx.LogContext, source string, target string) bool {
//	// TODO implement me
//	panic("implement me")
//}
//
// func (s *ServiceImpl) CanReach(context ctx.MoveContext, source string, target string) bool {
//	// TODO implement me
//	panic("implement me")
//}
