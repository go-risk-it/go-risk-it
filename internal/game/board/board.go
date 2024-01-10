package board

import (
	"github.com/tomfran/go-risk-it/internal/db"
	"go.uber.org/zap"
)

type Service struct {
	q   *db.Queries
	log *zap.SugaredLogger
}

type Region struct {
	ExternalReference int    `json:"id"`
	Name              string `json:"name"`
	ContinentId       int    `json:"continent_id"`
}

type Continent struct {
	ExternalReference int    `json:"id"`
	Name              string `json:"name"`
	BonusTroops       int    `json:"bonus_troops"`
}

type Border struct {
	FirstRegionId  int `json:"first_region_id"`
	SecondRegionId int `json:"second_region_id"`
}

type Board struct {
	Regions    []Region    `json:"regions"`
	Continents []Continent `json:"continents"`
	Borders    []Border    `json:"borders"`
}

func NewBoardService(queries *db.Queries, logger *zap.SugaredLogger) *Service {
	return &Service{q: queries, log: logger}
}

func (service *Service) PersistBoard(board *Board) error {
	// TODO: validate board
	// persist continents
	return nil
}
