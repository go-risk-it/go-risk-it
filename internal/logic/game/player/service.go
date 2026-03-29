package player

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
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
	slog.InfoContext(ctx, "fetching player state")

	result, err := querier.GetPlayersState(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	slog.InfoContext(ctx, "got player state")

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

	slog.InfoContext(ctx, "got players")

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

	slog.InfoContext(ctx, "got current player", "player", nil)

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

	slog.InfoContext(ctx, "got next player", "player", result)

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
	slog.InfoContext(ctx, "creating players", "players", players)

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

	slog.InfoContext(ctx, "created players", "players", players)

	result, err := querier.GetPlayersByGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players by game: %w", err)
	}

	return result, nil
}
