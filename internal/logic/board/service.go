package board

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"go.uber.org/zap"
)

type Service interface {
	FetchFromFile() (*Board, error)
}

type ServiceImpl struct {
	querier db.Querier
	log     *zap.SugaredLogger
}

type Region struct {
	ExternalReference string `json:"id"`
	Name              string `json:"name"`
	Continent         string `json:"continent"`
}

type Continent struct {
	ExternalReference string `json:"id"`
	Name              string `json:"name"`
	BonusTroops       int    `json:"bonusTroops"`
}

type Border struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type Board struct {
	Regions    []Region    `json:"layers"`
	Continents []Continent `json:"continents"`
	Borders    []Border    `json:"links"`
}

func NewService(q db.Querier, logger *zap.SugaredLogger) *ServiceImpl {
	return &ServiceImpl{querier: q, log: logger}
}

func (s *ServiceImpl) FetchFromFile() (*Board, error) {
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
