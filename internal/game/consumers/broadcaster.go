package consumers

import (
	"encoding/json"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	gameconfig "github.com/go-risk-it/go-risk-it/internal/game/config"
	"github.com/go-risk-it/go-risk-it/internal/game/consumers/converter"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/snapshot"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

// messageDispatcher sends a WS message to either a single player (WriteMessage)
// or all players (Broadcast).
type messageDispatcher func(ctx.GameContext, json.RawMessage)

// GameStateBroadcaster consumes game events from the bus and publishes state
// updates over WebSocket connections.
type GameStateBroadcaster struct {
	writer            Writer
	presence          Presence
	lifecycle         Lifecycle
	snapshotService   snapshot.Service
	missionController *MissionController
	moveLogController *MoveLogController
	historyConfig     gameconfig.HistoryConfig
}

// NewGameStateBroadcaster creates a broadcaster with narrow WS dependencies and
// domain services. Panics if historyConfig.Size <= 0.
func NewGameStateBroadcaster(
	writer Writer,
	presence Presence,
	lifecycle Lifecycle,
	snapshotService snapshot.Service,
	missionController *MissionController,
	moveLogController *MoveLogController,
	historyConfig gameconfig.HistoryConfig,
) *GameStateBroadcaster {
	if historyConfig.Size <= 0 {
		panic("HistoryConfig.Size must be > 0")
	}

	return &GameStateBroadcaster{
		writer:            writer,
		presence:          presence,
		lifecycle:         lifecycle,
		snapshotService:   snapshotService,
		missionController: missionController,
		moveLogController: moveLogController,
		historyConfig:     historyConfig,
	}
}

// Register subscribes the broadcaster's handlers to game events on the bus.
func (p *GameStateBroadcaster) Register(sub eventbus.Subscriber) {
	gameevt.OnGameEvent[*gameevt.MoveExecuted](sub, p.handleMoveExecuted)
	gameevt.OnGameEvent[*gameevt.PhaseTransitioned](sub, p.handlePhaseTransitioned)
	gameevt.OnGameEvent[*gameevt.GameCompleted](sub, p.handleGameCompleted)
	gameevt.OnGameEvent[*gameevt.PlayerConnected](sub, p.handlePlayerConnected)
}

func (p *GameStateBroadcaster) handleMoveExecuted(
	gameCtx ctx.GameContext,
	event *gameevt.MoveExecuted,
) {
	eventbus.SafeOp(gameCtx, "fetchAndPublishPublicState", func() {
		fetchAndPublishPublicState(gameCtx, p.snapshotService, p.presence, p.writer.Broadcast)
	})

	eventbus.SafeOp(gameCtx, "publishPrivateStates", func() {
		p.publishPrivateStates(gameCtx)
	})

	eventbus.SafeOp(gameCtx, "publishMoveLog", func() {
		p.publishMoveLog(gameCtx, event)
	})
}

func (p *GameStateBroadcaster) handlePlayerConnected(
	gameCtx ctx.GameContext,
	_ *gameevt.PlayerConnected,
) {
	eventbus.SafeOp(gameCtx, "fetchAndPublishPublicState", func() {
		fetchAndPublishPublicState(gameCtx, p.snapshotService, p.presence, p.writer.WriteMessage)
	})

	eventbus.SafeOp(gameCtx, "publishPrivateStateToPlayer", func() {
		p.publishPrivateStateToPlayer(gameCtx)
	})

	eventbus.SafeOp(gameCtx, "publishMoveHistory", func() {
		p.publishMoveHistory(gameCtx)
	})
}

func (p *GameStateBroadcaster) handlePhaseTransitioned(
	gameCtx ctx.GameContext,
	_ *gameevt.PhaseTransitioned,
) {
	eventbus.SafeOp(gameCtx, "fetchAndPublishPublicState", func() {
		fetchAndPublishPublicState(gameCtx, p.snapshotService, p.presence, p.writer.Broadcast)
	})

	eventbus.SafeOp(gameCtx, "publishPrivateStates", func() {
		p.publishPrivateStates(gameCtx)
	})
}

func (p *GameStateBroadcaster) handleGameCompleted(
	gameCtx ctx.GameContext,
	_ *gameevt.GameCompleted,
) {
	eventbus.SafeOp(gameCtx, "removeGame", func() {
		p.lifecycle.RemoveGame(gameCtx)
	})
}

func fetchAndPublishPublicState(
	gameCtx ctx.GameContext,
	snapshotService snapshot.Service,
	presence Presence,
	dispatch messageDispatcher,
) {
	snap, err := snapshotService.GetPublicSnapshot(gameCtx)
	if err != nil {
		observe.Error(gameCtx, err, "failed to get public snapshot")

		return
	}

	connectedPlayers := presence.GetConnectedPlayers(gameCtx)

	msgs, err := converter.ConvertPublicSnapshot(snap, connectedPlayers)
	if err != nil {
		observe.Error(gameCtx, err, "failed to convert public snapshot")

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
			observe.Error(gameCtx, err, "failed to build message",
				attribute.String("type", string(item.msgType)))

			return
		}

		dispatch(gameCtx, msg)
	}
}

func (p *GameStateBroadcaster) publishPrivateStates(gameCtx ctx.GameContext) {
	snapshots, err := p.snapshotService.GetPrivateSnapshotsByUser(gameCtx)
	if err != nil {
		observe.Error(gameCtx, err, "failed to get private snapshots")

		return
	}

	missionResolver := BuildMissionResolver(p.missionController)

	for userID, snap := range snapshots {
		msgs, err := converter.ConvertPrivateSnapshot(gameCtx, snap, missionResolver)
		if err != nil {
			observe.Error(gameCtx, err, "failed to convert private snapshot",
				attribute.String("user_id", userID))

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
				observe.Error(playerCtx, err, "failed to build private message",
					attribute.String("type", string(item.msgType)))

				continue
			}

			p.writer.WriteMessage(playerCtx, msg)
		}
	}
}

func (p *GameStateBroadcaster) publishMoveLog(
	gameCtx ctx.GameContext,
	event *gameevt.MoveExecuted,
) {
	history, err := p.moveLogController.ConvertMoveLogs(
		gameCtx,
		[]sqlc.GameMoveLog{event.MoveLog},
	)
	if err != nil {
		observe.Error(gameCtx, err, "failed to convert move log")

		return
	}

	msg, err := messaging.BuildMessage(messaging.MoveHistoryType, history)
	if err != nil {
		observe.Error(gameCtx, err, "failed to build move history message")

		return
	}

	p.writer.Broadcast(gameCtx, msg)
}

func (p *GameStateBroadcaster) publishPrivateStateToPlayer(gameCtx ctx.GameContext) {
	snapshots, err := p.snapshotService.GetPrivateSnapshotsByUser(gameCtx)
	if err != nil {
		observe.Error(gameCtx, err, "failed to get private snapshots")

		return
	}

	userID := gameCtx.UserID()

	snap, ok := snapshots[userID]
	if !ok {
		observe.Error(gameCtx, fmt.Errorf("no snapshot for %s", userID),
			"no private snapshot for connecting player",
			attribute.String("user_id", userID))

		return
	}

	missionResolver := BuildMissionResolver(p.missionController)

	msgs, err := converter.ConvertPrivateSnapshot(gameCtx, snap, missionResolver)
	if err != nil {
		observe.Error(gameCtx, err, "failed to convert private snapshot",
			attribute.String("user_id", userID))

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
			observe.Error(gameCtx, err, "failed to build private message",
				attribute.String("type", string(item.msgType)))

			continue
		}

		p.writer.WriteMessage(gameCtx, msg)
	}
}

func (p *GameStateBroadcaster) publishMoveHistory(gameCtx ctx.GameContext) {
	history, err := p.moveLogController.GetMoveLogs(gameCtx, p.historyConfig.Size)
	if err != nil {
		observe.Error(gameCtx, err, "failed to get move history")

		return
	}

	msg, err := messaging.BuildMessage(messaging.MoveHistoryType, history)
	if err != nil {
		observe.Error(gameCtx, err, "failed to build move history message")

		return
	}

	p.writer.WriteMessage(gameCtx, msg)
}
