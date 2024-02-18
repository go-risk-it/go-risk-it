package game

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/api/game/message/request"
	gameApi "github.com/tomfran/go-risk-it/internal/api/game/message/response"
	sqlc "github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"go.uber.org/zap"
)

type Controller interface {
	GetGameState(request request.GameStateRequest) (gameApi.GameStateResponse, error)
	GetFullState(request request.FullStateRequest) (gameApi.FullStateResponse, error)
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
	request request.GameStateRequest,
) (gameApi.GameStateResponse, error) {
	return gameApi.GameStateResponse{GameID: 1, Players: []gameApi.Player{}, CurrentTurn: 0}, nil
}

func (c *ControllerImpl) GetFullState(
	request request.FullStateRequest,
) (gameApi.FullStateResponse, error) {
	ctx := context.Background()

	gameState, err := c.gameService.GetGameState(ctx, request.GameID)
	if err != nil {
		return gameApi.FullStateResponse{}, fmt.Errorf("failed to get game state: %w", err)
	}

	players, err := c.playerService.GetPlayers(ctx, gameState.ID)
	if err != nil {
		return gameApi.FullStateResponse{}, fmt.Errorf("failed to get players: %w", err)
	}

	return gameApi.FullStateResponse{
		GameState: gameApi.GameStateResponse{
			GameID:       gameState.ID,
			Players:      convertPlayers(players),
			CurrentTurn:  gameState.CurrentTurn,
			CurrentPhase: string(gameState.CurrentPhase),
		},
		BoardState: gameApi.BoardStateResponse{
			Regions: []gameApi.Region{},
		},
	}, nil
}

func convertPlayers(players []sqlc.Player) []gameApi.Player {
	result := make([]gameApi.Player, len(players))
	for i, p := range players {
		result[i] = convertPlayer(p)
	}

	return result
}

func convertPlayer(player sqlc.Player) gameApi.Player {
	return gameApi.Player{
		PlayerID:  player.UserID,
		TurnIndex: player.TurnIndex,
	}
}
