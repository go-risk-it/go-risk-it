package controller

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/api/game/message"
	"github.com/tomfran/go-risk-it/internal/api/game/rest/request"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"go.uber.org/zap"
)

type GameController interface {
	CreateGame(ctx context.Context, request request.CreateGame) (int64, error)
	GetGameState(ctx context.Context, gameID int64) (message.GameState, error)
}

type GameControllerImpl struct {
	log           *zap.SugaredLogger
	gameService   game.Service
	playerService player.Service
	boardService  board.Service
}

func NewGameController(
	log *zap.SugaredLogger,
	gameService game.Service,
	boardService board.Service,
	playerService player.Service,
) *GameControllerImpl {
	return &GameControllerImpl{
		log:           log,
		gameService:   gameService,
		boardService:  boardService,
		playerService: playerService,
	}
}

func (c *GameControllerImpl) CreateGame(
	ctx context.Context, request request.CreateGame,
) (int64, error) {
	gameBoard, err := c.boardService.FetchFromFile()
	if err != nil {
		return -1, fmt.Errorf("failed to fetch board from file: %w", err)
	}

	gameID, err := c.gameService.CreateGameWithTx(ctx, gameBoard, request.Players)
	if err != nil {
		return -1, fmt.Errorf("failed to create game: %w", err)
	}

	return gameID, nil
}

func (c *GameControllerImpl) GetGameState(
	ctx context.Context, gameID int64,
) (message.GameState, error) {
	gameState, err := c.gameService.GetGameState(ctx, gameID)
	if err != nil {
		return message.GameState{}, fmt.Errorf("failed to get game state: %w", err)
	}

	return message.GameState{
		GameID:       gameState.ID,
		CurrentTurn:  gameState.Turn,
		CurrentPhase: string(gameState.Phase),
	}, nil
}
