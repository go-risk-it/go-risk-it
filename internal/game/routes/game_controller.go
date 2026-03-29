package routes

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/game/api/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/game/commands"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/creation"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

// GameController handles game creation and game summary queries.
// It implements [commands.Handler] so the kernel router can dispatch
// cross-module CreateGame commands from the lobby.
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

// HandleCreateGame implements commands.Handler for cross-module dispatch.
//
//nolint:contextcheck // cross-module boundary narrows context
func (c *GameController) HandleCreateGame(
	rawCtx context.Context,
	cmd commands.CreateGame,
) (commands.CreateGameResult, error) {
	userCtx, ok := rawCtx.(ctx.UserContext)
	if !ok {
		return commands.CreateGameResult{}, errors.New("HandleCreateGame requires UserContext")
	}

	regions, err := c.boardService.GetBoardRegions(userCtx)
	if err != nil {
		return commands.CreateGameResult{}, fmt.Errorf("failed to get board regions: %w", err)
	}

	players := make([]player.Player, len(cmd.Players))
	for i, p := range cmd.Players {
		players[i] = player.Player{
			UserID: p.UserID,
			Name:   p.Name,
		}
	}

	gameID, err := c.creationService.CreateGame(userCtx, regions, players)
	if err != nil {
		return commands.CreateGameResult{}, fmt.Errorf("failed to create game: %w", err)
	}

	return commands.CreateGameResult{GameID: gameID}, nil
}

// CreateGame handles the HTTP-facing game creation request (used by routes).
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

	gameID, err := c.creationService.CreateGame(ctx, regions, players)
	if err != nil {
		return -1, fmt.Errorf("failed to create game: %w", err)
	}

	return gameID, nil
}

// GetUserGames returns a summary of the user's games.
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

// Verify interface compliance.
var _ commands.Handler = (*GameController)(nil)
