package player

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

type Service interface {
	CreatePlayersQ(
		ctx ctx.GameContext,
		querier db.Querier,
		gameID int64,
		players []request.Player,
	) (
		[]sqlc.GamePlayer,
		error,
	)
	GetPlayersState(ctx ctx.GameContext) ([]sqlc.GetPlayersStateRow, error)
	GetPlayersStateQ(ctx ctx.GameContext, querier db.Querier) ([]sqlc.GetPlayersStateRow, error)
	GetPlayersQ(ctx ctx.GameContext, querier db.Querier) ([]sqlc.GamePlayer, error)
	GetCurrentPlayerQ(ctx ctx.GameContext, querier db.Querier) (sqlc.GamePlayer, error)
	GetNextPlayerQ(ctx ctx.GameContext, querier db.Querier) (sqlc.GamePlayer, error)
}

type ServiceImpl struct {
	querier db.Querier
}

var _ Service = (*ServiceImpl)(nil)

func NewService(querier db.Querier) *ServiceImpl {
	return &ServiceImpl{querier: querier}
}

func (s *ServiceImpl) GetPlayersState(ctx ctx.GameContext) ([]sqlc.GetPlayersStateRow, error) {
	return s.GetPlayersStateQ(ctx, s.querier)
}

func (s *ServiceImpl) GetPlayersStateQ(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.GetPlayersStateRow, error) {
	ctx.Log().Infow("fetching player state")

	result, err := querier.GetPlayersState(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	ctx.Log().Infow("got player state")

	return result, nil
}

func (s *ServiceImpl) GetPlayersQ(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.GamePlayer, error) {
	result, err := querier.GetPlayersByGame(ctx, ctx.GameID())
	if err != nil {
		return result, fmt.Errorf("failed to get players: %w", err)
	}

	ctx.Log().Infow("got players")

	return result, nil
}

func (s *ServiceImpl) GetCurrentPlayerQ(
	ctx ctx.GameContext,
	querier db.Querier,
) (sqlc.GamePlayer, error) {
	result, err := querier.GetCurrentPlayer(ctx, ctx.GameID())
	if err != nil {
		return sqlc.GamePlayer{}, fmt.Errorf("failed to get current player: %w", err)
	}

	ctx.Log().Infow("got current player", "player", nil)

	return result, nil
}

func (s *ServiceImpl) GetNextPlayerQ(
	ctx ctx.GameContext,
	querier db.Querier,
) (sqlc.GamePlayer, error) {
	result, err := querier.GetNextPlayer(ctx, ctx.GameID())
	if err != nil {
		return sqlc.GamePlayer{}, fmt.Errorf("failed to get next player: %w", err)
	}

	ctx.Log().Infow("got next player", "player", result)

	return result, nil
}

func (s *ServiceImpl) CreatePlayersQ(
	ctx ctx.GameContext,
	querier db.Querier,
	gameID int64,
	players []request.Player,
) ([]sqlc.GamePlayer, error) {
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
