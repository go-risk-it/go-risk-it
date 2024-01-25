package board

import (
	"github.com/tomfran/go-risk-it/internal/db"
	"go.uber.org/zap"
)

type Service interface{}

type ServiceImpl struct {
	q   *db.Queries
	log *zap.SugaredLogger
}

type Region struct {
	ExternalReference int    `json:"id"`
	Name              string `json:"name"`
	ContinentID       int    `json:"continentId"`
}

type Continent struct {
	ExternalReference int    `json:"id"`
	Name              string `json:"name"`
	BonusTroops       int    `json:"bonusTroops"`
}

type Border struct {
	FirstRegionID  int `json:"firstRegionId"`
	SecondRegionID int `json:"secondRegionId"`
}

type Board struct {
	Regions    []Region    `json:"regions"`
	Continents []Continent `json:"continents"`
	Borders    []Border    `json:"borders"`
}

func NewBoardService(queries *db.Queries, logger *zap.SugaredLogger) *ServiceImpl {
	return &ServiceImpl{q: queries, log: logger}
}
