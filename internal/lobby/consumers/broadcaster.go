package consumers

import (
	"encoding/json"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
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
	lobbyCtx ctx.LobbyContext,
	_ *lobbyevt.LobbyStateChanged,
) {
	p.fetchAndPublish(lobbyCtx, p.writer.Broadcast)
}

func (p *LobbyStateBroadcaster) onPlayerConnected(
	lobbyCtx ctx.LobbyContext,
	_ *lobbyevt.LobbyPlayerConnected,
) {
	p.fetchAndPublish(lobbyCtx, p.writer.WriteMessage)
}

// messageDispatcher sends a WS message to either a single player (WriteMessage)
// or all players (Broadcast).
type messageDispatcher func(ctx.LobbyContext, json.RawMessage)

func (p *LobbyStateBroadcaster) fetchAndPublish(
	lobbyCtx ctx.LobbyContext,
	dispatch messageDispatcher,
) {
	var msg json.RawMessage

	var fetchOk bool

	eventbus.SafeOp(lobbyCtx, "fetchLobbyState", func() {
		lobbyState, err := p.stateController.GetLobbyState(lobbyCtx)
		if err != nil {
			observe.Error(lobbyCtx, err, "failed to get lobby state")

			return
		}

		built, err := messaging.BuildMessage(messaging.LobbyStateType, lobbyState)
		if err != nil {
			observe.Error(lobbyCtx, err, "failed to build lobby state message")

			return
		}

		msg = built
		fetchOk = true
	})

	if !fetchOk {
		return
	}

	eventbus.SafeOp(lobbyCtx, "dispatchLobbyState", func() {
		dispatch(lobbyCtx, msg)
	})
}
