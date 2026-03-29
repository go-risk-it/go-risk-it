package routes

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/commands"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/router"
	"github.com/go-risk-it/go-risk-it/internal/lobby/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/lobby/logic/start"
)

// StartController orchestrates game start by dispatching a CreateGame command
// through the kernel router. This decouples the lobby module from the game
// module — the router is the only cross-module dispatch point.
type StartController struct {
	router       *router.Router
	startService start.Service
}

func NewStartController(
	router *router.Router,
	startService start.Service,
) *StartController {
	return &StartController{
		router:       router,
		startService: startService,
	}
}

func (c *StartController) StartGame(ctx ctx.LobbyContext) error {
	canStartLobby, err := c.startService.CanStartLobby(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if lobby can be started: %w", err)
	}

	if !canStartLobby {
		return errors.New("lobby cannot be started")
	}

	lobbyPlayers, err := c.startService.GetLobbyPlayers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get lobby players: %w", err)
	}

	result, err := c.router.Route(ctx, buildCreateGameCommand(lobbyPlayers))
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	createGameResult, ok := result.(commands.CreateGameResult)
	if !ok {
		return fmt.Errorf("unexpected result type from router: %T", result)
	}

	if err := c.startService.MarkLobbyAsStarted(ctx, createGameResult.GameID); err != nil {
		return fmt.Errorf("failed to mark lobby as started: %w", err)
	}

	slog.InfoContext(ctx, "lobby started", "game_id", createGameResult.GameID)

	return nil
}

func buildCreateGameCommand(players []sqlc.GetLobbyPlayersRow) commands.CreateGame {
	cmd := commands.CreateGame{
		Players: make([]commands.Player, len(players)),
	}

	for idx, player := range players {
		cmd.Players[idx] = commands.Player{
			Name:   player.Name,
			UserID: player.UserID,
		}
	}

	return cmd
}
