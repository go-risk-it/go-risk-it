package game

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/api/game/message"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"go.uber.org/zap"
)

type Controller interface {
	GetGameState(ctx context.Context, gameID int64) (message.GameState, error)
}

type ControllerImpl struct {
	log           *zap.SugaredLogger
	gameService   game.Service
	playerService player.Service
	boardService  board.Service
}

func New(
	log *zap.SugaredLogger,
	gameService game.Service,
	boardService board.Service,
	playerService player.Service,
) *ControllerImpl {
	return &ControllerImpl{
		log:           log,
		gameService:   gameService,
		boardService:  boardService,
		playerService: playerService,
	}
}

func (c *ControllerImpl) GetGameState(
	ctx context.Context, gameID int64,
) (message.GameState, error) {
	gameState, err := c.gameService.GetGameState(ctx, gameID)
	if err != nil {
		return message.GameState{}, fmt.Errorf("failed to get game state: %w", err)
	}

	return message.GameState{
		GameID:       gameState.ID,
		CurrentTurn:  gameState.CurrentTurn,
		CurrentPhase: string(gameState.CurrentPhase),
	}, nil
}
