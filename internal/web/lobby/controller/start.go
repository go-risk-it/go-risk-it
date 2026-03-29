package controller

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/api/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/lobby/logic/start"
)

type GameCreator interface {
	CreateGame(ctx ctx.UserContext, request request.CreateGame) (int64, error)
}

type StartController struct {
	gameCreator  GameCreator
	startService start.Service
}

func NewStartController(
	gameCreator GameCreator,
	startService start.Service,
) *StartController {
	return &StartController{
		gameCreator:  gameCreator,
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

	gameID, err := c.gameCreator.CreateGame(ctx, buildCreateGameRequest(lobbyPlayers))
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	if err := c.startService.MarkLobbyAsStarted(ctx, gameID); err != nil {
		return fmt.Errorf("failed to mark lobby as started: %w", err)
	}

	slog.InfoContext(ctx, "lobby started", "game_id", gameID)

	return nil
}

func buildCreateGameRequest(players []sqlc.GetLobbyPlayersRow) request.CreateGame {
	res := request.CreateGame{
		Players: make([]request.Player, len(players)),
	}

	for idx, player := range players {
		res.Players[idx] = request.Player{
			Name:   player.Name,
			UserID: player.UserID,
		}
	}

	return res
}
