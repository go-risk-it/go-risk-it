package controller

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/start"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
)

type StartController interface {
	StartGame(ctx ctx.LobbyContext) error
}

type StartControllerImpl struct {
	gameController controller.GameController
	startService   start.Service
}

var _ StartController = (*StartControllerImpl)(nil)

func NewStartController(
	gameController controller.GameController,
	startService start.Service,
) *StartControllerImpl {
	return &StartControllerImpl{
		gameController: gameController,
		startService:   startService,
	}
}

func (c *StartControllerImpl) StartGame(ctx ctx.LobbyContext) error {
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

	gameID, err := c.gameController.CreateGame(ctx, buildCreateGameRequest(lobbyPlayers))
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	if err := c.startService.MarkLobbyAsStarted(ctx, gameID); err != nil {
		return fmt.Errorf("failed to mark lobby as started: %w", err)
	}

	ctx.Log().Infow("lobby started", "game_id", gameID)

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
