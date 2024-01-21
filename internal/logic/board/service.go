package board

import (
	"github.com/tomfran/go-risk-it/internal/db"
	"go.uber.org/zap"
)

type Service interface {
	PersistBoard(board *Board) error
}

type ServiceImpl struct {
	q   *db.Queries
	log *zap.SugaredLogger
}

type Region struct {
	ExternalReference int    `json:"id"`
	Name              string `json:"name"`
	ContinentID       int    `json:"continent_id"`
}

type Continent struct {
	ExternalReference int    `json:"id"`
	Name              string `json:"name"`
	BonusTroops       int    `json:"bonus_troops"`
}

type Border struct {
	FirstRegionID  int `json:"first_region_id"`
	SecondRegionID int `json:"second_region_id"`
}

type Board struct {
	Regions    []Region    `json:"regions"`
	Continents []Continent `json:"continents"`
	Borders    []Border    `json:"borders"`
}

func NewBoardService(queries *db.Queries, logger *zap.SugaredLogger) *ServiceImpl {
	return &ServiceImpl{q: queries, log: logger}
}
