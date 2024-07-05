package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/message"
	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/creation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type GameController interface {
	CreateGame(ctx ctx.UserContext, request request.CreateGame) (int64, error)
	GetGameState(ctx ctx.GameContext) (message.GameState, error)
}

type GameControllerImpl struct {
	boardService    board.Service
	creationService creation.Service
	gameService     state.Service
}

var _ GameController = (*GameControllerImpl)(nil)

func NewGameController(
	boardService board.Service,
	creationService creation.Service,
	gameService state.Service,
) *GameControllerImpl {
	return &GameControllerImpl{
		boardService:    boardService,
		creationService: creationService,
		gameService:     gameService,
	}
}

func (c *GameControllerImpl) CreateGame(
	ctx ctx.UserContext, request request.CreateGame,
) (int64, error) {
	regions, err := c.boardService.GetBoardRegions(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to get board regions: %w", err)
	}

	gameID, err := c.creationService.CreateGameWithTx(ctx, regions, request.Players)
	if err != nil {
		return -1, fmt.Errorf("failed to create game: %w", err)
	}

	return gameID, nil
}

func (c *GameControllerImpl) GetGameState(ctx ctx.GameContext) (message.GameState, error) {
	ctx.Log().Infow("fetching game state")

	gameState, err := c.gameService.GetGameState(ctx)
	if err != nil {
		return message.GameState{}, fmt.Errorf("failed to get game state: %w", err)
	}

	return message.GameState{
		ID:           gameState.ID,
		CurrentTurn:  gameState.CurrentTurn,
		CurrentPhase: convertPhase(gameState.CurrentPhase),
	}, nil
}

func convertPhase(phase sqlc.PhaseType) message.Phase {
	switch phase {
	case sqlc.PhaseTypeCARDS:
		return message.Cards
	case sqlc.PhaseTypeDEPLOY:
		return message.Deploy
	case sqlc.PhaseTypeATTACK:
		return message.Attack
	case sqlc.PhaseTypeCONQUER:
		return message.Conquer
	default:
		return ""
	}
}
