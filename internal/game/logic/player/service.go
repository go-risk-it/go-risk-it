package player

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
)

type Service interface {
	CreatePlayers(
		ctx ctx.GameContext,
		querier db.Querier,
		gameID int64,
		players []Player,
	) (
		[]sqlc.GamePlayer,
		error,
	)
	GetPlayersState(ctx ctx.GameContext) ([]sqlc.GetPlayersStateRow, error)
	GetPlayersStateWithQuerier(
		ctx ctx.GameContext,
		querier db.Querier,
	) ([]sqlc.GetPlayersStateRow, error)
	GetPlayers(ctx ctx.GameContext, querier db.Querier) ([]sqlc.GamePlayer, error)
	GetCurrentPlayer(ctx ctx.GameContext, querier db.Querier) (sqlc.GamePlayer, error)
	GetNextPlayer(ctx ctx.GameContext, querier db.Querier) (sqlc.GamePlayer, error)
}

type service struct {
	querier     db.Querier
	gameService state.Service
}

var _ Service = (*service)(nil)

func NewService(
	querier db.Querier,
	gameService state.Service,
) Service {
	return &service{
		querier:     querier,
		gameService: gameService,
	}
}

func (s *service) GetPlayersState(ctx ctx.GameContext) ([]sqlc.GetPlayersStateRow, error) {
	return s.GetPlayersStateWithQuerier(ctx, s.querier)
}

func (s *service) GetPlayersStateWithQuerier(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.GetPlayersStateRow, error) {
	result, err := querier.GetPlayersState(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	return result, nil
}

func (s *service) GetPlayers(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.GamePlayer, error) {
	result, err := querier.GetPlayersByGame(ctx, ctx.GameID())
	if err != nil {
		return result, fmt.Errorf("failed to get players: %w", err)
	}

	return result, nil
}

func (s *service) GetCurrentPlayer(
	ctx ctx.GameContext,
	querier db.Querier,
) (sqlc.GamePlayer, error) {
	result, err := querier.GetCurrentPlayer(ctx, ctx.GameID())
	if err != nil {
		return sqlc.GamePlayer{}, fmt.Errorf("failed to get current player: %w", err)
	}

	return result, nil
}

func (s *service) GetNextPlayer(
	ctx ctx.GameContext,
	querier db.Querier,
) (sqlc.GamePlayer, error) {
	nextTurn, err := s.getNextTurn(ctx, querier)
	if err != nil {
		return sqlc.GamePlayer{}, fmt.Errorf("failed to get players state: %w", err)
	}

	result, err := querier.GetPlayerAtTurnIndex(ctx, sqlc.GetPlayerAtTurnIndexParams{
		GameID: ctx.GameID(),
		Turn:   nextTurn,
	})
	if err != nil {
		return sqlc.GamePlayer{}, fmt.Errorf("failed to get next player: %w", err)
	}

	return result, nil
}

func (s *service) getNextTurn(
	ctx ctx.GameContext,
	querier db.Querier,
) (int64, error) {
	gameState, err := s.gameService.GetGameStateWithQuerier(ctx, querier)
	if err != nil {
		return -1, fmt.Errorf("failed to get game state: %w", err)
	}

	turn := gameState.Turn

	playersState, err := s.GetPlayersStateWithQuerier(ctx, querier)
	if err != nil {
		return -1, fmt.Errorf("failed to get players state: %w", err)
	}

	turn++

	players := int64(len(playersState))
	for playersState[turn%players].RegionCount == 0 {
		turn++
	}

	return turn, nil
}

func (s *service) CreatePlayers(
	ctx ctx.GameContext,
	querier db.Querier,
	gameID int64,
	players []Player,
) ([]sqlc.GamePlayer, error) {
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

	result, err := querier.GetPlayersByGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players by game: %w", err)
	}

	return result, nil
}
