package board

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
)

type Service interface {
	GetBoard() (*Board, error)
}

type ServiceImpl struct {
	log        *zap.SugaredLogger
	boardCache *Board
}

var _ Service = (*ServiceImpl)(nil)

func NewService(logger *zap.SugaredLogger) *ServiceImpl {
	return &ServiceImpl{log: logger, boardCache: nil}
}

func (s *ServiceImpl) GetBoard() (*Board, error) {
	if s.boardCache != nil {
		return s.boardCache, nil
	}

	board, err := s.fetchFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	s.boardCache = board

	return board, nil
}

func (s *ServiceImpl) fetchFromFile() (*Board, error) {
	data, err := os.ReadFile("map.json")
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	board := &Board{}

	err = json.Unmarshal(data, board)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	s.log.Debugw("Read board from file", "board", board)

	return board, nil
}
