package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/creation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type GameController struct {
	boardService    board.Service
	creationService creation.Service
	gameService     state.Service
}

func NewGameController(
	boardService board.Service,
	creationService creation.Service,
	gameService state.Service,
) *GameController {
	return &GameController{
		boardService:    boardService,
		creationService: creationService,
		gameService:     gameService,
	}
}

func (c *GameController) CreateGame(
	ctx ctx.UserContext, req request.CreateGame,
) (int64, error) {
	regions, err := c.boardService.GetBoardRegions(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to get board regions: %w", err)
	}

	players := make([]player.Player, len(req.Players))
	for i, p := range req.Players {
		players[i] = player.Player{
			UserID: p.UserID,
			Name:   p.Name,
		}
	}

	gameID, err := c.creationService.CreateGameWithTx(ctx, regions, players)
	if err != nil {
		return -1, fmt.Errorf("failed to create game: %w", err)
	}

	return gameID, nil
}

func (c *GameController) GetUserGames(ctx ctx.UserContext) (response.Games, error) {
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
