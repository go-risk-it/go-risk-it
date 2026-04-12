package routes

import (
	"fmt"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/rest/response"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/lobby/events"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic/start"
	"go.opentelemetry.io/otel/attribute"
)

const startTimeout = 10 * time.Second

// StartController orchestrates game start by emitting a CreateGameRequested
// event and awaiting the asynchronous game creation result via PendingStarts.
type StartController struct {
	bus          bus.Publisher
	startService start.Service
	pending      *start.PendingStarts
}

func NewStartController(
	bus bus.Publisher,
	startService start.Service,
	pending *start.PendingStarts,
) *StartController {
	return &StartController{
		bus:          bus,
		startService: startService,
		pending:      pending,
	}
}

func (c *StartController) StartGame(lobbyCtx ctx.LobbyContext) (response.StartGame, error) {
	canStartLobby, err := c.startService.CanStartLobby(lobbyCtx)
	if err != nil {
		return response.StartGame{}, fmt.Errorf("failed to check if lobby can be started: %w", err)
	}

	if !canStartLobby {
		return response.StartGame{}, domainerrors.NewValidationError("lobby cannot be started")
	}

	lobbyID := lobbyCtx.LobbyID()

	resultChan, err := c.pending.Register(lobbyID)
	if err != nil {
		return response.StartGame{}, fmt.Errorf("failed to register pending start: %w", err)
	}

	lobbyPlayers, err := c.startService.GetLobbyPlayers(lobbyCtx)
	if err != nil {
		c.pending.Cancel(lobbyID)

		return response.StartGame{}, fmt.Errorf("failed to get lobby players: %w", err)
	}

	players := make([]lobbyevt.LobbyPlayer, len(lobbyPlayers))
	for i, p := range lobbyPlayers {
		players[i] = lobbyevt.LobbyPlayer{
			UserID: p.UserID,
			Name:   p.Name,
		}
	}

	c.bus.Emit(lobbyCtx, lobbyevt.NewCreateGameRequested(
		lobbyID,
		lobbyCtx.UserID(),
		time.Now(),
		players,
	))

	gameID, err := c.pending.Await(lobbyCtx, lobbyID, resultChan, startTimeout)
	if err != nil {
		return response.StartGame{}, fmt.Errorf("failed to start game: %w", err)
	}

	if err := c.startService.MarkLobbyAsStarted(lobbyCtx, gameID); err != nil {
		return response.StartGame{}, fmt.Errorf("failed to mark lobby as started: %w", err)
	}

	observe.Info(lobbyCtx, "lobby started", attribute.Int64("game_id", gameID))

	return response.StartGame{GameID: gameID}, nil
}
