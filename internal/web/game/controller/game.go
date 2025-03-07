package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/creation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type GameController interface {
	CreateGame(ctx ctx.UserContext, request request.CreateGame) (int64, error)
	GetUserGames(ctx ctx.UserContext) (response.Games, error)
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

func (c *GameControllerImpl) GetUserGames(ctx ctx.UserContext) (response.Games, error) {
	userGames, err := c.gameService.GetUserGames(ctx)
	if err != nil {
		return response.Games{}, fmt.Errorf("failed to get user games: %w", err)
	}

	result := make([]response.Game, 0)

	for _, gameID := range userGames {
		result = append(result, response.Game{
			ID: gameID,
		})
	}

	return response.Games{
		Games: result,
	}, nil
}
