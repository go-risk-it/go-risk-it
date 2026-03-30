package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	gameconfig "github.com/go-risk-it/go-risk-it/internal/game/config"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/publisher/converter"
	"github.com/go-risk-it/go-risk-it/internal/game/snapshot"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

// messageDispatcher sends a WS message to either a single player (WriteMessage)
// or all players (Broadcast).
type messageDispatcher func(ctx.GameContext, json.RawMessage)

const publisherTracerName = "go-risk-it-publisher"

// GameStatePublisher consumes game events from the bus and publishes state
// updates over WebSocket connections. Each handler performs sequential ordered
// delivery within a single goroutine (the bus dispatches each handler in its
// own goroutine). Sub-operations are wrapped in safeOp for independent panic
// recovery -- a panic in one sub-operation does not prevent the others from
// executing.
type GameStatePublisher struct {
	writer            Writer
	presence          Presence
	lifecycle         Lifecycle
	snapshotService   snapshot.Service
	missionController *MissionController
	moveLogController *MoveLogController
	historyConfig     gameconfig.HistoryConfig
	metrics           *metrics.InfraMetrics
}

// NewGameStatePublisher creates a publisher with narrow WS dependencies and
// domain services. Panics if historyConfig.Size <= 0.
func NewGameStatePublisher(
	writer Writer,
	presence Presence,
	lifecycle Lifecycle,
	snapshotService snapshot.Service,
	missionController *MissionController,
	moveLogController *MoveLogController,
	historyConfig gameconfig.HistoryConfig,
	met *metrics.InfraMetrics,
) *GameStatePublisher {
	if historyConfig.Size <= 0 {
		panic("HistoryConfig.Size must be > 0")
	}

	return &GameStatePublisher{
		writer:            writer,
		presence:          presence,
		lifecycle:         lifecycle,
		snapshotService:   snapshotService,
		missionController: missionController,
		moveLogController: moveLogController,
		historyConfig:     historyConfig,
		metrics:           met,
	}
}

// Register subscribes the publisher's handlers to game events on the bus.
func (p *GameStatePublisher) Register(bus eventbus.Bus) {
	gameevt.OnGameEvent[*gameevt.MoveExecuted](bus, p.handleMoveExecuted)
	gameevt.OnGameEvent[*gameevt.PhaseTransitioned](bus, p.handlePhaseTransitioned)
	gameevt.OnGameEvent[*gameevt.GameCompleted](bus, p.handleGameCompleted)
	gameevt.OnGameEvent[*gameevt.PlayerConnected](bus, p.handlePlayerConnected)
}

// handleMoveExecuted broadcasts public state, then private states per player,
// then the move log entry to all connected players.
func (p *GameStatePublisher) handleMoveExecuted(
	gameCtx ctx.GameContext,
	event *gameevt.MoveExecuted,
) {
	safeOp(gameCtx, "fetchAndPublishPublicState", p.metrics, func() {
		fetchAndPublishPublicState(gameCtx, p.snapshotService, p.presence, p.writer.Broadcast)
	})

	safeOp(gameCtx, "publishPrivateStates", p.metrics, func() {
		p.publishPrivateStates(gameCtx)
	})

	safeOp(gameCtx, "publishMoveLog", p.metrics, func() {
		p.publishMoveLog(gameCtx, event)
	})
}

// handlePlayerConnected sends full game state to the newly connected player:
// public state via WriteMessage, then private state for this player, then
// recent move history.
func (p *GameStatePublisher) handlePlayerConnected(
	gameCtx ctx.GameContext,
	_ *gameevt.PlayerConnected,
) {
	safeOp(gameCtx, "fetchAndPublishPublicState", p.metrics, func() {
		fetchAndPublishPublicState(gameCtx, p.snapshotService, p.presence, p.writer.WriteMessage)
	})

	safeOp(gameCtx, "publishPrivateStateToPlayer", p.metrics, func() {
		p.publishPrivateStateToPlayer(gameCtx)
	})

	safeOp(gameCtx, "publishMoveHistory", p.metrics, func() {
		p.publishMoveHistory(gameCtx)
	})
}

// handlePhaseTransitioned broadcasts updated public and private state after a
// phase advance (e.g., attack -> reinforce). This is separate from MoveExecuted
// because the advancement service emits PhaseTransitioned directly without a
// MoveExecuted event.
func (p *GameStatePublisher) handlePhaseTransitioned(
	gameCtx ctx.GameContext,
	_ *gameevt.PhaseTransitioned,
) {
	safeOp(gameCtx, "fetchAndPublishPublicState", p.metrics, func() {
		fetchAndPublishPublicState(gameCtx, p.snapshotService, p.presence, p.writer.Broadcast)
	})

	safeOp(gameCtx, "publishPrivateStates", p.metrics, func() {
		p.publishPrivateStates(gameCtx)
	})
}

// handleGameCompleted cleans up the game's connection tracking.
func (p *GameStatePublisher) handleGameCompleted(
	gameCtx ctx.GameContext,
	_ *gameevt.GameCompleted,
) {
	safeOp(gameCtx, "removeGame", p.metrics, func() {
		p.lifecycle.RemoveGame(gameCtx)
	})
}

// safeOp runs action with a child span and duration metric recording. On panic
// it records the error on the span and logs the recovered value and stack trace.
// This is a sequential wrapper (not a goroutine) -- the bus already owns
// goroutine lifecycle.
func safeOp(
	parent context.Context,
	name string,
	met *metrics.InfraMetrics,
	action func(),
) {
	ctx, span := otel.GetTracerProvider().
		Tracer(publisherTracerName).
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

// fetchAndPublishPublicState fetches the public snapshot, converts it into typed
// DTOs, serializes them into WS message envelopes, and dispatches each using the
// provided dispatcher.
func fetchAndPublishPublicState(
	gameCtx ctx.GameContext,
	snapshotService snapshot.Service,
	presence Presence,
	dispatch messageDispatcher,
) {
	snap, err := snapshotService.GetPublicSnapshot(gameCtx)
	if err != nil {
		slog.ErrorContext(gameCtx, "failed to get public snapshot", "error", err)

		return
	}

	connectedPlayers := presence.GetConnectedPlayers(gameCtx)

	msgs, err := converter.ConvertPublicSnapshot(snap, connectedPlayers)
	if err != nil {
		slog.ErrorContext(gameCtx, "failed to convert public snapshot", "error", err)

		return
	}

	for _, item := range []struct {
		msgType messaging.Type
		payload any
	}{
		{messaging.GameStateType, msgs.GameState},
		{messaging.BoardStateType, msgs.BoardState},
		{messaging.PlayerStateType, msgs.PlayerState},
	} {
		msg, err := messaging.BuildMessage(item.msgType, item.payload)
		if err != nil {
			slog.ErrorContext(
				gameCtx,
				"failed to build message",
				"type",
				item.msgType,
				"error",
				err,
			)

			return
		}

		dispatch(gameCtx, msg)
	}
}

// publishPrivateStates fetches per-player private snapshots, converts each
// into typed DTOs, serializes them into WS message envelopes, and writes
// them to the corresponding player's connection.
func (p *GameStatePublisher) publishPrivateStates(gameCtx ctx.GameContext) {
	snapshots, err := p.snapshotService.GetPrivateSnapshotsByUser(gameCtx)
	if err != nil {
		slog.ErrorContext(gameCtx, "failed to get private snapshots", "error", err)

		return
	}

	missionResolver := BuildMissionResolver(p.missionController)

	for userID, snap := range snapshots {
		msgs, err := converter.ConvertPrivateSnapshot(gameCtx, snap, missionResolver)
		if err != nil {
			slog.ErrorContext(
				gameCtx,
				"failed to convert private snapshot",
				"userID", userID,
				"error", err,
			)

			continue
		}

		playerCtx := ctx.WithGameID(
			kernelctx.WithUserID(kernelctx.WithSpan(gameCtx, gameCtx.Span()), userID),
			gameCtx.GameID(),
		)

		for _, item := range []struct {
			msgType messaging.Type
			payload any
		}{
			{messaging.CardStateType, msgs.CardState},
			{messaging.MissionStateType, msgs.MissionState},
		} {
			msg, err := messaging.BuildMessage(item.msgType, item.payload)
			if err != nil {
				slog.ErrorContext(playerCtx, "failed to build private message",
					"type", item.msgType, "error", err)

				continue
			}

			p.writer.WriteMessage(playerCtx, msg)
		}
	}
}

// publishMoveLog converts the move log from the event and broadcasts it to all
// connected players.
func (p *GameStatePublisher) publishMoveLog(
	gameCtx ctx.GameContext,
	event *gameevt.MoveExecuted,
) {
	history, err := p.moveLogController.ConvertMoveLogs(
		gameCtx,
		[]sqlc.GameMoveLog{event.MoveLog},
	)
	if err != nil {
		slog.ErrorContext(gameCtx, "failed to convert move log", "error", err)

		return
	}

	msg, err := messaging.BuildMessage(messaging.MoveHistoryType, history)
	if err != nil {
		slog.ErrorContext(gameCtx, "failed to build move history message", "error", err)

		return
	}

	p.writer.Broadcast(gameCtx, msg)
}

// publishPrivateStateToPlayer fetches per-player private snapshots, extracts the
// connecting player's snapshot, converts it into typed DTOs, serializes them into
// WS message envelopes, and writes them to the connecting player's connection.
func (p *GameStatePublisher) publishPrivateStateToPlayer(gameCtx ctx.GameContext) {
	snapshots, err := p.snapshotService.GetPrivateSnapshotsByUser(gameCtx)
	if err != nil {
		slog.ErrorContext(gameCtx, "failed to get private snapshots", "error", err)

		return
	}

	userID := gameCtx.UserID()

	snap, ok := snapshots[userID]
	if !ok {
		slog.ErrorContext(gameCtx, "no private snapshot for connecting player", "userID", userID)

		return
	}

	missionResolver := BuildMissionResolver(p.missionController)

	msgs, err := converter.ConvertPrivateSnapshot(gameCtx, snap, missionResolver)
	if err != nil {
		slog.ErrorContext(
			gameCtx,
			"failed to convert private snapshot",
			"userID", userID,
			"error", err,
		)

		return
	}

	for _, item := range []struct {
		msgType messaging.Type
		payload any
	}{
		{messaging.CardStateType, msgs.CardState},
		{messaging.MissionStateType, msgs.MissionState},
	} {
		msg, err := messaging.BuildMessage(item.msgType, item.payload)
		if err != nil {
			slog.ErrorContext(gameCtx, "failed to build private message",
				"type", item.msgType, "error", err)

			continue
		}

		p.writer.WriteMessage(gameCtx, msg)
	}
}

// publishMoveHistory fetches the recent move history and writes it to the
// connecting player.
func (p *GameStatePublisher) publishMoveHistory(gameCtx ctx.GameContext) {
	history, err := p.moveLogController.GetMoveLogs(gameCtx, p.historyConfig.Size)
	if err != nil {
		slog.ErrorContext(gameCtx, "failed to get move history", "error", err)

		return
	}

	msg, err := messaging.BuildMessage(messaging.MoveHistoryType, history)
	if err != nil {
		slog.ErrorContext(gameCtx, "failed to build move history message", "error", err)

		return
	}

	p.writer.WriteMessage(gameCtx, msg)
}
