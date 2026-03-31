package consumers

import (
	"encoding/json"
	"fmt"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/messaging"
	lobbyctx "github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/lobby/events"
)

// LobbyStateBroadcaster consumes lobby events from the bus and publishes state
// updates over WebSocket connections.
type LobbyStateBroadcaster struct {
	writer          Writer
	stateController *StateController
}

// NewLobbyStateBroadcaster creates a broadcaster with narrow WS and controller
// dependencies.
func NewLobbyStateBroadcaster(
	writer Writer,
	stateController *StateController,
) *LobbyStateBroadcaster {
	return &LobbyStateBroadcaster{
		writer:          writer,
		stateController: stateController,
	}
}

// Register subscribes the broadcaster's handlers to lobby events on the bus.
func (p *LobbyStateBroadcaster) Register(sub eventbus.Subscriber) {
	lobbyevt.OnLobbyEvent(sub, p.onStateChanged)
	lobbyevt.OnLobbyEvent(sub, p.onPlayerConnected)
}

func (p *LobbyStateBroadcaster) onStateChanged(
	lCtx lobbyctx.LobbyContext,
	_ *lobbyevt.LobbyStateChanged,
) {
	p.fetchAndPublish(lCtx, p.writer.Broadcast)
}

func (p *LobbyStateBroadcaster) onPlayerConnected(
	lCtx lobbyctx.LobbyContext,
	_ *lobbyevt.LobbyPlayerConnected,
) {
	p.fetchAndPublish(lCtx, p.writer.WriteMessage)
}

// messageDispatcher sends a WS message to either a single player (WriteMessage)
// or all players (Broadcast).
type messageDispatcher func(lobbyctx.LobbyContext, json.RawMessage)

func (p *LobbyStateBroadcaster) fetchAndPublish(
	lCtx lobbyctx.LobbyContext,
	dispatch messageDispatcher,
) {
	var msg json.RawMessage

	var fetchOk bool

	LobbySafeOp(lCtx, "fetchLobbyState", func(lCtx lobbyctx.LobbyContext) error {
		lobbyState, err := p.stateController.GetLobbyState(lCtx)
		if err != nil {
			return fmt.Errorf("failed to get lobby state: %w", err)
		}

		built, err := messaging.BuildMessage(messaging.LobbyStateType, lobbyState)
		if err != nil {
			return fmt.Errorf("failed to build lobby state message: %w", err)
		}

		msg = built
		fetchOk = true

		return nil
	})

	if !fetchOk {
		return
	}

	LobbySafeOp(lCtx, "dispatchLobbyState", func(lCtx lobbyctx.LobbyContext) error {
		dispatch(lCtx, msg)

		return nil
	})
}
