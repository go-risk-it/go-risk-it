package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	lobbyevt "github.com/go-risk-it/go-risk-it/internal/events/lobby"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

const lobbyPublisherTracerName = "go-risk-it-lobby-publisher"

// LobbyStatePublisher consumes lobby events from the bus and publishes state
// updates over WebSocket connections. It replaces the channel-based fetcher
// pattern with direct service calls, since the bus already manages goroutine
// lifecycle and context detachment.
type LobbyStatePublisher struct {
	writer          ws.Writer
	stateController *controller.StateController
	metrics         *metrics.InfraMetrics
}

// NewLobbyStatePublisher creates a publisher with narrow WS and controller
// dependencies.
func NewLobbyStatePublisher(
	writer ws.Writer,
	stateController *controller.StateController,
	met *metrics.InfraMetrics,
) *LobbyStatePublisher {
	return &LobbyStatePublisher{
		writer:          writer,
		stateController: stateController,
		metrics:         met,
	}
}

// Register subscribes the publisher's handlers to lobby events on the bus.
func (p *LobbyStatePublisher) Register(bus eventbus.Bus) {
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

	safeOp(lobbyCtx, "fetchLobbyState", p.metrics, func() {
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

	safeOp(lobbyCtx, "dispatchLobbyState", p.metrics, func() {
		dispatch(lobbyCtx, msg)
	})
}

// safeOp runs action with a child span and duration metric recording. On panic
// it records the error on the span and logs the recovered value and stack trace.
// This is a sequential wrapper (not a goroutine) — the bus already owns
// goroutine lifecycle.
func safeOp(
	parent context.Context,
	name string,
	met *metrics.InfraMetrics,
	action func(),
) {
	ctx, span := otel.GetTracerProvider().
		Tracer(lobbyPublisherTracerName).
		Start(parent, "consumer."+name)
	defer span.End()

	start := time.Now()

	defer func() {
		elapsed := time.Since(start).Seconds()

		if met != nil {
			met.EventHandlerDuration.Record(ctx, elapsed,
				metric.WithAttributes(attribute.String("handler", name)))
		}

		if recovered := recover(); recovered != nil {
			span.RecordError(fmt.Errorf("panic in %s: %v", name, recovered))
			span.SetStatus(codes.Error, "panic")

			slog.ErrorContext(ctx, "panic in consumer operation",
				"operation", name,
				"error", recovered,
				"stack", string(debug.Stack()),
			)
		}
	}()

	action()
}
