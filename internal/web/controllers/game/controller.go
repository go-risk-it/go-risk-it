package game

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/api/game/message"
	"github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"go.uber.org/zap"
)

type Controller interface {
	GetGameState(gameID int64) (message.GameState, error)
	GetFullState(gameID int64) (message.FullState, error)
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
	gameID int64,
) (message.GameState, error) {
	return message.GameState{GameID: gameID, Players: []message.Player{}, CurrentTurn: 0}, nil
}

func (c *ControllerImpl) GetFullState(
	gameID int64,
) (message.FullState, error) {
	ctx := context.Background()

	gameState, err := c.gameService.GetGameState(ctx, gameID)
	if err != nil {
		return message.FullState{}, fmt.Errorf("failed to get game state: %w", err)
	}

	players, err := c.playerService.GetPlayers(ctx, gameState.ID)
	if err != nil {
		return message.FullState{}, fmt.Errorf("failed to get players: %w", err)
	}

	return message.FullState{
		GameState: message.GameState{
			GameID:       gameState.ID,
			Players:      convertPlayers(players),
			CurrentTurn:  gameState.CurrentTurn,
			CurrentPhase: string(gameState.CurrentPhase),
		},
		BoardState: message.BoardState{
			Regions: []message.Region{},
		},
	}, nil
}

func convertPlayers(players []sqlc.Player) []message.Player {
	result := make([]message.Player, len(players))
	for i, p := range players {
		result[i] = convertPlayer(p)
	}

	return result
}

func convertPlayer(player sqlc.Player) message.Player {
	return message.Player{
		PlayerID:  player.UserID,
		TurnIndex: player.TurnIndex,
	}
}
