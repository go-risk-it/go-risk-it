package player

import (
	"context"

	"github.com/tomfran/go-risk-it/internal/db"
	"go.uber.org/zap"
)

type Service interface {
	CreatePlayers(ctx context.Context, q db.Querier, gameId int64, users []string) (
		[]db.Player,
		error,
	)
}

type ServiceImpl struct {
	log *zap.SugaredLogger
}

func NewPlayersService(log *zap.SugaredLogger) *ServiceImpl {
	return &ServiceImpl{log: log}
}

func (s *ServiceImpl) CreatePlayers(
	ctx context.Context,
	q db.Querier,
	gameID int64,
	users []string,
) ([]db.Player, error) {
	s.log.Infow("creating players", "gameId", gameID, "users", users)
	playersParams := make([]db.InsertPlayersParams, 0, len(users))
	for _, user := range users {
		playersParams = append(playersParams, db.InsertPlayersParams{GameID: gameID, UserID: user})
	}
	if _, err := q.InsertPlayers(ctx, playersParams); err != nil {
		s.log.Errorw("failed to insert players", "error", err)
		return nil, err
	}
	s.log.Infow("created players", "gameId", gameID, "users", users)
	return q.GetPlayersByGameId(ctx, gameID)
}
