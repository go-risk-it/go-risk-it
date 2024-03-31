package player

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
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
	GetPlayersQ(ctx context.Context, querier db.Querier, gameID int64) (
		[]sqlc.Player,
		error,
	)

	DecreaseDeployableTroopsQ(
		ctx context.Context,
		querier db.Querier,
		player *sqlc.Player,
		troops int64,
	) error
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
	return s.GetPlayersQ(ctx, s.querier, gameID)
}

func (s *ServiceImpl) GetPlayersQ(ctx context.Context, querier db.Querier, gameID int64) (
	[]sqlc.Player,
	error,
) {
	result, err := querier.GetPlayersByGame(ctx, gameID)
	if err != nil {
		return result, fmt.Errorf("failed to get players: %w", err)
	}

	s.log.Infow("got players", "gameID", gameID)

	return result, nil
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
				GameID:           gameID,
				UserID:           user,
				TurnIndex:        turnIndex,
				DeployableTroops: 0,
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

func (s *ServiceImpl) DecreaseDeployableTroopsQ(
	ctx context.Context,
	querier db.Querier,
	player *sqlc.Player,
	troops int64,
) error {
	s.log.Infow("decreasing deployable troops", "player", player, "troops", troops)

	if player.DeployableTroops < troops {
		return fmt.Errorf("player does not have enough troops to deploy")
	}

	err := querier.DecreaseDeployableTroops(ctx, sqlc.DecreaseDeployableTroopsParams{
		ID:               player.ID,
		DeployableTroops: troops,
	})
	if err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	s.log.Infow("decreased deployable troops", "player", player, "troops", troops)

	return nil
}
