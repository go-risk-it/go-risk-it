package board

import (
	"github.com/tomfran/go-risk-it/internal/data/db"
	"go.uber.org/zap"
)

type Service interface{}

type ServiceImpl struct {
	querier db.Querier
	log     *zap.SugaredLogger
}

type Region struct {
	ExternalReference string `json:"id"`
	Name              string `json:"name"`
	ContinentID       int    `json:"continentId"`
}

type Continent struct {
	ExternalReference string `json:"id"`
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

func NewService(q db.Querier, logger *zap.SugaredLogger) *ServiceImpl {
	return &ServiceImpl{querier: q, log: logger}
}
