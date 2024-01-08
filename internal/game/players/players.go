package players

import (
	"context"
	"github.com/tomfran/go-risk-it/internal/db"
)

type PlayersService struct {
	q *db.Queries
}

func NewPlayersService(queries *db.Queries) *PlayersService {
	return &PlayersService{q: queries}
}

func (service *PlayersService) GetPlayers() ([]db.Player, error) {
	players, err := service.q.GetPlayers(context.Background())
	if err != nil {
		return nil, err
	}
	return players, nil
}
