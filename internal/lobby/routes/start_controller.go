package routes

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/commands"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/lobby/logic/start"
)

// StartController orchestrates game start by dispatching a CreateGame command
// directly to the game module's command handler.
type StartController struct {
	gameHandler  commands.Handler
	startService start.Service
}

func NewStartController(
	gameHandler commands.Handler,
	startService start.Service,
) *StartController {
	return &StartController{
		gameHandler:  gameHandler,
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

	createGameResult, err := c.gameHandler.HandleCreateGame(
		ctx,
		buildCreateGameCommand(lobbyPlayers),
	)
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
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
