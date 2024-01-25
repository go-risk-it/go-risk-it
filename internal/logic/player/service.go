package player

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/db"
	"go.uber.org/zap"
)

type Service interface {
	CreatePlayers(ctx context.Context, q db.Querier, gameID int64, users []string) (
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
	querier db.Querier,
	gameID int64,
	users []string,
) ([]db.Player, error) {
	s.log.Infow("creating players", "gameID", gameID, "users", users)

	turnIndex := int64(0)
	playersParams := make([]db.InsertPlayersParams, 0, len(users))

	for _, user := range users {
		playersParams = append(
			playersParams,
			db.InsertPlayersParams{GameID: gameID, UserID: user, TurnIndex: turnIndex},
		)
		turnIndex += 1
	}

	if _, err := querier.InsertPlayers(ctx, playersParams); err != nil {
		return nil, fmt.Errorf("failed to insert players: %w", err)
	}

	s.log.Infow("created players", "gameId", gameID, "users", users)

	players, err := querier.GetPlayersByGameId(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players by game ID: %w", err)
	}

	return players, nil
}
