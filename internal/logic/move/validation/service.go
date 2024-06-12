package validation

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/player"
	"go.uber.org/zap"
)

type Service interface {
	Validate(
		ctx context.Context,
		querier db.Querier,
		gameID int64,
		userID string,
		phase sqlc.Phase,
	) error
}

type ServiceImpl struct {
	log           *zap.SugaredLogger
	gameService   game.Service
	playerService player.Service
}

func NewService(
	log *zap.SugaredLogger,
	gameService game.Service,
	playerService player.Service,
) *ServiceImpl {
	return &ServiceImpl{log: log, gameService: gameService, playerService: playerService}
}

func (s *ServiceImpl) Validate(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	userID string,
	phase sqlc.Phase,
) error {
	players, err := s.playerService.GetPlayersQ(ctx, querier, gameID)
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	thisPlayer := extractPlayerFrom(userID, players)
	if thisPlayer == nil {
		return fmt.Errorf("player is not in game")
	}

	if err := s.checkTurn(
		ctx,
		querier,
		gameID,
		int64(len(players)),
		thisPlayer.TurnIndex,
		phase); err != nil {
		return fmt.Errorf("turn check failed: %w", err)
	}

	return nil
}

func (s *ServiceImpl) checkTurn(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	playersInGame int64,
	playerTurn int64,
	phase sqlc.Phase,
) error {
	gameState, err := s.gameService.GetGameStateQ(ctx, querier, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	// move to specific services
	if gameState.Phase != phase {
		return fmt.Errorf("game is not in %s phase", phase)
	}

	if gameState.Turn%playersInGame != playerTurn {
		return fmt.Errorf("it is not the player's turn")
	}

	return nil
}

func extractPlayerFrom(userID string, players []sqlc.Player) *sqlc.Player {
	for _, p := range players {
		if p.UserID == userID {
			return &p
		}
	}

	return nil
}
