package player

import (
	"context"

	"github.com/tomfran/go-risk-it/internal/db"
	"go.uber.org/zap"
)

type Service struct {
	log *zap.SugaredLogger
}

func NewPlayersService(log *zap.SugaredLogger) *Service {
	return &Service{log: log}
}

func (s *Service) CreatePlayers(ctx context.Context, q *db.Queries, gameId int64, users []string) ([]db.Player, error) {
	s.log.Infow("creating players", "gameId", gameId, "users", users)
	var playersParams []db.InsertPlayersParams
	for _, user := range users {
		playersParams = append(playersParams, db.InsertPlayersParams{GameID: gameId, UserID: user})
	}
	if _, err := q.InsertPlayers(ctx, playersParams); err != nil {
		s.log.Errorw("failed to insert players", "error", err)
		return nil, err
	}
	s.log.Infow("created players", "gameId", gameId, "users", users)
	return q.GetPlayersByGameId(ctx, gameId)
}
