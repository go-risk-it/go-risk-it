package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/message"
	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
)

type GameController interface {
	CreateGame(ctx ctx.UserContext, request request.CreateGame) (int64, error)
	GetGameState(ctx ctx.GameContext) (message.GameState, error)
}

type GameControllerImpl struct {
	gameService   gamestate.Service
	playerService player.Service
	boardService  board.Service
}

func NewGameController(
	gameService gamestate.Service,
	boardService board.Service,
	playerService player.Service,
) *GameControllerImpl {
	return &GameControllerImpl{
		gameService:   gameService,
		boardService:  boardService,
		playerService: playerService,
	}
}

func (c *GameControllerImpl) CreateGame(
	ctx ctx.UserContext, request request.CreateGame,
) (int64, error) {
	gameBoard, err := c.boardService.GetBoard()
	if err != nil {
		return -1, fmt.Errorf("failed to fetch board from file: %w", err)
	}

	gameID, err := c.gameService.CreateGameWithTx(ctx, gameBoard, request.Players)
	if err != nil {
		return -1, fmt.Errorf("failed to create game: %w", err)
	}

	return gameID, nil
}

func (c *GameControllerImpl) GetGameState(ctx ctx.GameContext) (message.GameState, error) {
	ctx.Log().Infow("fetching game state")

	gameState, err := c.gameService.GetGameState(ctx, ctx.GameID())
	if err != nil {
		return message.GameState{}, fmt.Errorf("failed to get game state: %w", err)
	}

	return message.GameState{
		GameID:           gameState.ID,
		CurrentTurn:      gameState.Turn,
		CurrentPhase:     string(gameState.Phase),
		DeployableTroops: gameState.DeployableTroops,
	}, nil
}
