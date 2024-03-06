package player

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/data/db"
	sqlc "github.com/tomfran/go-risk-it/internal/data/sqlc"
	"go.uber.org/zap"
)

type Service interface {
	CreatePlayers(ctx context.Context, querier db.Querier, gameID int64, users []string) (
		[]sqlc.Player,
		error,
	)
	GetPlayers(ctx context.Context, gameID int64) (
		[]sqlc.Player,
		error,
	)
}

type ServiceImpl struct {
	log     *zap.SugaredLogger
	querier db.Querier
}

func NewService(log *zap.SugaredLogger, querier db.Querier) *ServiceImpl {
	return &ServiceImpl{log: log, querier: querier}
}

func (s *ServiceImpl) GetPlayers(ctx context.Context, gameID int64) (
	[]sqlc.Player,
	error,
) {
	players, err := s.querier.GetPlayersByGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	return players, nil
}

func (s *ServiceImpl) CreatePlayers(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	users []string,
) ([]sqlc.Player, error) {
	s.log.Infow("creating players", "gameID", gameID, "users", users)

	turnIndex := int64(0)
	playersParams := make([]sqlc.InsertPlayersParams, 0, len(users))

	for _, user := range users {
		playersParams = append(
			playersParams,
			sqlc.InsertPlayersParams{
				GameID:         gameID,
				UserID:         user,
				TurnIndex:      turnIndex,
				TroopsToDeploy: 0,
			},
		)
		turnIndex += 1
	}

	if _, err := querier.InsertPlayers(ctx, playersParams); err != nil {
		return nil, fmt.Errorf("failed to insert players: %w", err)
	}

	s.log.Infow("created players", "gameId", gameID, "users", users)

	players, err := querier.GetPlayersByGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players by game ID: %w", err)
	}

	return players, nil
}
