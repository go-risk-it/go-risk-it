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
	GetFullState(ctx context.Context, gameID int64) (message.FullState, error)
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
	return message.GameState{GameID: gameID, CurrentTurn: 0}, nil
}

func (c *ControllerImpl) GetFullState(
	ctx context.Context, gameID int64,
) (message.FullState, error) {
	gameState, err := c.gameService.GetGameState(ctx, gameID)
	if err != nil {
		return message.FullState{}, fmt.Errorf("failed to get game state: %w", err)
	}

	return message.FullState{
		GameState: message.GameState{
			GameID:       gameState.ID,
			CurrentTurn:  gameState.CurrentTurn,
			CurrentPhase: string(gameState.CurrentPhase),
		},
		BoardState: message.BoardState{
			Regions: []message.Region{},
		},
	}, nil
}
