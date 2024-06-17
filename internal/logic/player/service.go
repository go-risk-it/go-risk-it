package player

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"go.uber.org/zap"
)

type Service interface {
	CreatePlayers(
		ctx context.Context,
		querier db.Querier,
		gameID int64,
		players []request.Player,
	) (
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
	players []request.Player,
) ([]sqlc.Player, error) {
	s.log.Infow("creating players", "gameID", gameID, "players", players)

	turnIndex := int64(0)
	playersParams := make([]sqlc.InsertPlayersParams, 0, len(players))

	for _, player := range players {
		playersParams = append(
			playersParams,
			sqlc.InsertPlayersParams{
				GameID:    gameID,
				UserID:    player.UserID,
				Name:      player.Name,
				TurnIndex: turnIndex,
			},
		)
		turnIndex += 1
	}

	if _, err := querier.InsertPlayers(ctx, playersParams); err != nil {
		return nil, fmt.Errorf("failed to insert players: %w", err)
	}

	s.log.Infow("created players", "gameId", gameID, "players", players)

	result, err := querier.GetPlayersByGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players by game UserID: %w", err)
	}

	return result, nil
}
