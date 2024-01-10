package player

import (
	"context"
	"github.com/tomfran/go-risk-it/internal/db"
	"go.uber.org/zap"
)

type Service struct {
	q   *db.Queries
	log *zap.SugaredLogger
}

func NewPlayersService(queries *db.Queries, log *zap.SugaredLogger) *Service {
	return &Service{q: queries, log: log}
}

func (s *Service) GetPlayersByGame() ([]db.Player, error) {
	return s.q.GetPlayersByGameId(context.Background(), 1)
}

func (s *Service) CreatePlayers(gameId int64, users []string) ([]db.Player, error) {
	s.log.Infow("creating players", "gameId", gameId, "users", users)
	var playersParams []db.InsertPlayersParams
	for _, user := range users {
		playersParams = append(playersParams, db.InsertPlayersParams{GameID: gameId, UserID: user})
	}
	if _, err := s.q.InsertPlayers(context.Background(), playersParams); err != nil {
		s.log.Errorw("failed to insert players", "error", err)
		return nil, err
	}
	s.log.Infow("created players", "gameId", gameId, "users", users)
	return s.q.GetPlayersByGameId(context.Background(), gameId)
}
