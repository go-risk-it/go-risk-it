package player

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

type Service interface {
	CreatePlayers(
		ctx ctx.GameContext,
		querier db.Querier,
		gameID int64,
		players []request.Player,
	) (
		[]sqlc.Player,
		error,
	)
	GetPlayersState(ctx ctx.GameContext) ([]sqlc.GetPlayersStateRow, error)
	GetPlayersQ(ctx ctx.GameContext, querier db.Querier) ([]sqlc.Player, error)
}

type ServiceImpl struct {
	querier db.Querier
}

var _ Service = (*ServiceImpl)(nil)

func NewService(querier db.Querier) *ServiceImpl {
	return &ServiceImpl{querier: querier}
}

func (s *ServiceImpl) GetPlayersState(ctx ctx.GameContext) ([]sqlc.GetPlayersStateRow, error) {
	ctx.Log().Infow("fetching player state")

	result, err := s.querier.GetPlayersState(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	ctx.Log().Infow("got player state")

	return result, nil
}

func (s *ServiceImpl) GetPlayersQ(ctx ctx.GameContext, querier db.Querier) ([]sqlc.Player, error) {
	result, err := querier.GetPlayersByGame(ctx, ctx.GameID())
	if err != nil {
		return result, fmt.Errorf("failed to get players: %w", err)
	}

	ctx.Log().Infow("got players")

	return result, nil
}

func (s *ServiceImpl) CreatePlayers(
	ctx ctx.GameContext,
	querier db.Querier,
	gameID int64,
	players []request.Player,
) ([]sqlc.Player, error) {
	ctx.Log().Infow("creating players", "players", players)

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

	ctx.Log().Infow("created players", "players", players)

	result, err := querier.GetPlayersByGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players by game: %w", err)
	}

	return result, nil
}
