package routes

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/game/api/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/creation"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

// GameController handles game creation and game summary queries.
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

// CreateGame handles the HTTP-facing game creation request (used by routes).
func (c *GameController) CreateGame(
	ctx ctx.UserContext, req request.CreateGame,
) (response.CreateGame, error) {
	regions, err := c.boardService.GetBoardRegions(ctx)
	if err != nil {
		return response.CreateGame{}, fmt.Errorf("failed to get board regions: %w", err)
	}

	players := make([]player.Player, len(req.Players))
	for i, p := range req.Players {
		players[i] = player.Player{
			UserID: p.UserID,
			Name:   p.Name,
		}
	}

	gameID, err := c.creationService.CreateGame(ctx, 0, regions, players)
	if err != nil {
		return response.CreateGame{}, fmt.Errorf("failed to create game: %w", err)
	}

	return response.CreateGame{GameID: gameID}, nil
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
