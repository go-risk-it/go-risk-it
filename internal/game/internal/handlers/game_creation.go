package handlers

import (
	"context"
	"errors"
	"time"

	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/creation"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/player"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/kernel/safego"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/lobby/events"
	"go.uber.org/fx"
)

// gameCreationHandler listens for CreateGameRequested events from the lobby
// module and orchestrates game creation. On success, the creation service
// emits GameCreated (with lobbyID). On failure, this handler emits
// GameCreationFailed so the lobby can resolve the pending start.
type gameCreationHandler struct {
	bus             eventbus.Publisher
	boardService    board.Service
	creationService creation.Service
}

// GameCreationHandlerParams holds the dependencies for the game creation handler.
type GameCreationHandlerParams struct {
	fx.In

	Sub             eventbus.Subscriber
	Bus             eventbus.Publisher
	BoardService    board.Service
	CreationService creation.Service
}

// RegisterGameCreationHandler subscribes the game creation handler to
// CreateGameRequested events from the lobby module.
func RegisterGameCreationHandler(params GameCreationHandlerParams) {
	h := &gameCreationHandler{
		bus:             params.Bus,
		boardService:    params.BoardService,
		creationService: params.CreationService,
	}

	eventbus.OnEvent(params.Sub, h.handleCreateGameRequested)
}

func (h *gameCreationHandler) handleCreateGameRequested(
	rawCtx context.Context,
	event *lobbyevt.CreateGameRequested,
) {
	safego.SafeOp(rawCtx, "game_creation", func(ctx context.Context) error {
		userCtx, ok := ctx.(kernelctx.UserContext)
		if !ok {
			reason := "context is not UserContext"
			observe.Error(ctx, errors.New(reason), "game creation handler: context assertion failed")
			h.bus.Emit(ctx, gameevt.NewGameCreationFailed(
				event.LobbyID(), time.Now(), reason,
			))

			return nil
		}

		regions, err := h.boardService.GetBoardRegions(userCtx)
		if err != nil {
			observe.Error(userCtx, err, "game creation handler: failed to get board regions")
			h.bus.Emit(userCtx, gameevt.NewGameCreationFailed(
				event.LobbyID(), time.Now(), err.Error(),
			))

			return nil
		}

		players := make([]player.Player, len(event.Players))
		for i, p := range event.Players {
			players[i] = player.Player{
				UserID: p.UserID,
				Name:   p.Name,
			}
		}

		_, err = h.creationService.CreateGame(userCtx, event.LobbyID(), regions, players)
		if err != nil {
			observe.Error(userCtx, err, "game creation handler: failed to create game")
			h.bus.Emit(userCtx, gameevt.NewGameCreationFailed(
				event.LobbyID(), time.Now(), err.Error(),
			))

			return nil
		}

		// GameCreated is emitted by the creation service itself (with lobbyID threaded through).

		return nil
	})
}
