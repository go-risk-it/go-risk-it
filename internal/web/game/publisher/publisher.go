package publisher

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime/debug"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/events"
	gameevt "github.com/go-risk-it/go-risk-it/internal/events/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/converter"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

// messageDispatcher sends a WS message to either a single player (WriteMessage)
// or all players (Broadcast).
type messageDispatcher func(ctx.GameContext, json.RawMessage)

// GameStatePublisher consumes game events from the bus and publishes state
// updates over WebSocket connections. Each handler performs sequential ordered
// delivery within a single goroutine (the bus dispatches each handler in its
// own goroutine). Sub-operations are wrapped in safeOp for independent panic
// recovery — a panic in one sub-operation does not prevent the others from
// executing.
type GameStatePublisher struct {
	writer            ws.Writer
	presence          ws.Presence
	lifecycle         ws.Lifecycle
	snapshotService   snapshot.Service
	missionController *controller.MissionController
	moveLogController *controller.MoveLogController
	historyConfig     config.HistoryConfig
}

// NewGameStatePublisher creates a publisher with narrow WS dependencies and
// domain services.
func NewGameStatePublisher(
	writer ws.Writer,
	presence ws.Presence,
	lifecycle ws.Lifecycle,
	snapshotService snapshot.Service,
	missionController *controller.MissionController,
	moveLogController *controller.MoveLogController,
	historyConfig config.HistoryConfig,
) *GameStatePublisher {
	return &GameStatePublisher{
		writer:            writer,
		presence:          presence,
		lifecycle:         lifecycle,
		snapshotService:   snapshotService,
		missionController: missionController,
		moveLogController: moveLogController,
		historyConfig:     historyConfig,
	}
}

// Register subscribes the publisher's handlers to game events on the bus.
func (p *GameStatePublisher) Register(bus events.Bus) {
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
	safeOp(gameCtx, "fetchAndPublishPublicState", func() {
		fetchAndPublishPublicState(gameCtx, p.snapshotService, p.presence, p.writer.Broadcast)
	})

	safeOp(gameCtx, "publishPrivateStates", func() {
		p.publishPrivateStates(gameCtx)
	})

	safeOp(gameCtx, "publishMoveLog", func() {
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
	safeOp(gameCtx, "fetchAndPublishPublicState", func() {
		fetchAndPublishPublicState(gameCtx, p.snapshotService, p.presence, p.writer.WriteMessage)
	})

	safeOp(gameCtx, "publishPrivateStateToPlayer", func() {
		p.publishPrivateStateToPlayer(gameCtx)
	})

	safeOp(gameCtx, "publishMoveHistory", func() {
		p.publishMoveHistory(gameCtx)
	})
}

// handlePhaseTransitioned broadcasts updated public and private state after a
// phase advance (e.g., attack → reinforce). This is separate from MoveExecuted
// because the advancement service emits PhaseTransitioned directly without a
// MoveExecuted event.
func (p *GameStatePublisher) handlePhaseTransitioned(
	gameCtx ctx.GameContext,
	_ *gameevt.PhaseTransitioned,
) {
	safeOp(gameCtx, "fetchAndPublishPublicState", func() {
		fetchAndPublishPublicState(gameCtx, p.snapshotService, p.presence, p.writer.Broadcast)
	})

	safeOp(gameCtx, "publishPrivateStates", func() {
		p.publishPrivateStates(gameCtx)
	})
}

// handleGameCompleted cleans up the game's connection tracking.
func (p *GameStatePublisher) handleGameCompleted(
	gameCtx ctx.GameContext,
	_ *gameevt.GameCompleted,
) {
	safeOp(gameCtx, "removeGame", func() {
		p.lifecycle.RemoveGame(gameCtx)
	})
}

// safeOp runs fn with panic recovery. On panic it logs the recovered value and
// stack trace and returns false. On normal return it forwards fn's bool result.
// This is a sequential wrapper (not a goroutine) — the bus already owns
// goroutine lifecycle.
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

// fetchAndPublishPublicState fetches the public snapshot, converts it into WS
// messages, and dispatches each using the provided dispatcher.
func fetchAndPublishPublicState(
	gameCtx ctx.GameContext,
	snapshotService snapshot.Service,
	presence ws.Presence,
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

	for _, msg := range []json.RawMessage{
		msgs.GameState,
		msgs.BoardState,
		msgs.PlayerState,
	} {
		dispatch(gameCtx, msg)
	}
}

// publishPrivateStates fetches per-player private snapshots, converts each
// into WS messages, and writes them to the corresponding player's connection.
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
			ctx.WithUserID(ctx.WithSpan(gameCtx, gameCtx.Span()), userID),
			gameCtx.GameID(),
		)

		for _, msg := range []json.RawMessage{
			msgs.CardState,
			msgs.MissionState,
		} {
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

	msg, err := message.BuildMessage(message.MoveHistory, history)
	if err != nil {
		slog.ErrorContext(gameCtx, "failed to build move history message", "error", err)

		return
	}

	p.writer.Broadcast(gameCtx, msg)
}

// publishPrivateStateToPlayer fetches per-player private snapshots, extracts the
// connecting player's snapshot, converts it into WS messages, and writes them to
// the connecting player's connection.
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

	for _, msg := range []json.RawMessage{
		msgs.CardState,
		msgs.MissionState,
	} {
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

	msg, err := message.BuildMessage(message.MoveHistory, history)
	if err != nil {
		slog.ErrorContext(gameCtx, "failed to build move history message", "error", err)

		return
	}

	p.writer.WriteMessage(gameCtx, msg)
}
