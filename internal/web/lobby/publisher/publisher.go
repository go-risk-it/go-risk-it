package publisher

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime/debug"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/events"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/events/lobby"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

// LobbyStatePublisher consumes lobby events from the bus and publishes state
// updates over WebSocket connections. It replaces the channel-based fetcher
// pattern with direct service calls, since the bus already manages goroutine
// lifecycle and context detachment.
type LobbyStatePublisher struct {
	writer          ws.Writer
	stateController *controller.StateController
}

// NewLobbyStatePublisher creates a publisher with narrow WS and controller
// dependencies.
func NewLobbyStatePublisher(
	writer ws.Writer,
	stateController *controller.StateController,
) *LobbyStatePublisher {
	return &LobbyStatePublisher{
		writer:          writer,
		stateController: stateController,
	}
}

// Register subscribes the publisher's handlers to lobby events on the bus.
func (p *LobbyStatePublisher) Register(bus events.Bus) {
	lobbyevt.OnLobbyEvent(bus, p.onStateChanged)
	lobbyevt.OnLobbyEvent(bus, p.onPlayerConnected)
}

// onStateChanged fetches the current lobby state and broadcasts it to all
// connected players.
func (p *LobbyStatePublisher) onStateChanged(
	lobbyCtx ctx.LobbyContext,
	_ *lobbyevt.LobbyStateChanged,
) {
	p.fetchAndPublish(lobbyCtx, p.writer.Broadcast)
}

// onPlayerConnected fetches the current lobby state and sends it to the
// newly connected player.
func (p *LobbyStatePublisher) onPlayerConnected(
	lobbyCtx ctx.LobbyContext,
	_ *lobbyevt.LobbyPlayerConnected,
) {
	p.fetchAndPublish(lobbyCtx, p.writer.WriteMessage)
}

// messageDispatcher sends a WS message to either a single player (WriteMessage)
// or all players (Broadcast).
type messageDispatcher func(ctx.LobbyContext, json.RawMessage)

// fetchAndPublish fetches the lobby state, builds a WS message, and dispatches
// it using the provided dispatcher. Each sub-operation is wrapped in safeOp for
// panic recovery — a panic in message building must not prevent future
// deliveries.
func (p *LobbyStatePublisher) fetchAndPublish(
	lobbyCtx ctx.LobbyContext,
	dispatch messageDispatcher,
) {
	var msg json.RawMessage

	var fetchOk bool

	safeOp(lobbyCtx, "fetch lobby state", func() {
		lobbyState, err := p.stateController.GetLobbyState(lobbyCtx)
		if err != nil {
			slog.ErrorContext(lobbyCtx, "failed to get lobby state", "error", err)

			return
		}

		built, err := message.BuildMessage(message.LobbyState, lobbyState)
		if err != nil {
			slog.ErrorContext(lobbyCtx, "failed to build lobby state message", "error", err)

			return
		}

		msg = built
		fetchOk = true
	})

	if !fetchOk {
		return
	}

	safeOp(lobbyCtx, "dispatch lobby state", func() {
		dispatch(lobbyCtx, msg)
	})
}

// safeOp runs fn with panic recovery. On panic it logs the recovered value and
// stack trace. This is a sequential wrapper (not a goroutine) — the bus already
// owns goroutine lifecycle.
func safeOp(c context.Context, name string, action func()) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(c, "panic in publisher sub-operation",
				"operation", name,
				"panic", r,
				"stack", string(debug.Stack()),
			)
		}
	}()

	action()
}
