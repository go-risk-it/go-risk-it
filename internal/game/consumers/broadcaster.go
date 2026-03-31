package consumers

import (
	"encoding/json"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	gameconfig "github.com/go-risk-it/go-risk-it/internal/game/config"
	"github.com/go-risk-it/go-risk-it/internal/game/consumers/converter"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/snapshot"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

// messageDispatcher sends a WS message to either a single player (WriteMessage)
// or all players (Broadcast).
type messageDispatcher func(gamectx.GameContext, json.RawMessage)

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
	gameCtx gamectx.GameContext,
	event *gameevt.MoveExecuted,
) {
	GameSafeOp(gameCtx, "fetchAndPublishPublicState", func(gameCtx gamectx.GameContext) error {
		return fetchAndPublishPublicState(
			gameCtx, p.snapshotService, p.presence, p.writer.Broadcast,
		)
	})

	GameSafeOp(gameCtx, "publishPrivateStates", func(gameCtx gamectx.GameContext) error {
		return p.publishPrivateStates(gameCtx)
	})

	GameSafeOp(gameCtx, "publishMoveLog", func(gameCtx gamectx.GameContext) error {
		return p.publishMoveLog(gameCtx, event)
	})
}

func (p *GameStateBroadcaster) handlePlayerConnected(
	gameCtx gamectx.GameContext,
	_ *gameevt.PlayerConnected,
) {
	GameSafeOp(gameCtx, "fetchAndPublishPublicState", func(gameCtx gamectx.GameContext) error {
		return fetchAndPublishPublicState(
			gameCtx, p.snapshotService, p.presence, p.writer.WriteMessage,
		)
	})

	GameSafeOp(gameCtx, "publishPrivateStateToPlayer", func(gameCtx gamectx.GameContext) error {
		return p.publishPrivateStateToPlayer(gameCtx)
	})

	GameSafeOp(gameCtx, "publishMoveHistory", func(gameCtx gamectx.GameContext) error {
		return p.publishMoveHistory(gameCtx)
	})
}

func (p *GameStateBroadcaster) handlePhaseTransitioned(
	gameCtx gamectx.GameContext,
	_ *gameevt.PhaseTransitioned,
) {
	GameSafeOp(gameCtx, "fetchAndPublishPublicState", func(gameCtx gamectx.GameContext) error {
		return fetchAndPublishPublicState(
			gameCtx, p.snapshotService, p.presence, p.writer.Broadcast,
		)
	})

	GameSafeOp(gameCtx, "publishPrivateStates", func(gameCtx gamectx.GameContext) error {
		return p.publishPrivateStates(gameCtx)
	})
}

func (p *GameStateBroadcaster) handleGameCompleted(
	gameCtx gamectx.GameContext,
	_ *gameevt.GameCompleted,
) {
	GameSafeOp(gameCtx, "removeGame", func(gameCtx gamectx.GameContext) error {
		p.lifecycle.RemoveGame(gameCtx)

		return nil
	})
}

func fetchAndPublishPublicState(
	gameCtx gamectx.GameContext,
	snapshotService snapshot.Service,
	presence Presence,
	dispatch messageDispatcher,
) error {
	snap, err := snapshotService.GetPublicSnapshot(gameCtx)
	if err != nil {
		return fmt.Errorf("failed to get public snapshot: %w", err)
	}

	connectedPlayers := presence.GetConnectedPlayers(gameCtx)

	msgs, err := converter.ConvertPublicSnapshot(snap, connectedPlayers)
	if err != nil {
		return fmt.Errorf("failed to convert public snapshot: %w", err)
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
			return fmt.Errorf("failed to build %s message: %w", item.msgType, err)
		}

		dispatch(gameCtx, msg)
	}

	return nil
}

func (p *GameStateBroadcaster) publishPrivateStates(gameCtx gamectx.GameContext) error {
	snapshots, err := p.snapshotService.GetPrivateSnapshotsByUser(gameCtx)
	if err != nil {
		return fmt.Errorf("failed to get private snapshots: %w", err)
	}

	missionResolver := BuildMissionResolver(p.missionController)

	for userID, snap := range snapshots {
		msgs, err := converter.ConvertPrivateSnapshot(gameCtx, snap, missionResolver)
		if err != nil {
			return fmt.Errorf("failed to convert private snapshot for %s: %w", userID, err)
		}

		playerCtx := gamectx.WithGameID(
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
				return fmt.Errorf("failed to build %s message for %s: %w",
					item.msgType, userID, err)
			}

			p.writer.WriteMessage(playerCtx, msg)
		}
	}

	return nil
}

func (p *GameStateBroadcaster) publishMoveLog(
	gameCtx gamectx.GameContext,
	event *gameevt.MoveExecuted,
) error {
	history, err := p.moveLogController.ConvertMoveLogs(
		gameCtx,
		[]sqlc.GameMoveLog{event.MoveLog},
	)
	if err != nil {
		return fmt.Errorf("failed to convert move log: %w", err)
	}

	msg, err := messaging.BuildMessage(messaging.MoveHistoryType, history)
	if err != nil {
		return fmt.Errorf("failed to build move history message: %w", err)
	}

	p.writer.Broadcast(gameCtx, msg)

	return nil
}

func (p *GameStateBroadcaster) publishPrivateStateToPlayer(gameCtx gamectx.GameContext) error {
	snapshots, err := p.snapshotService.GetPrivateSnapshotsByUser(gameCtx)
	if err != nil {
		return fmt.Errorf("failed to get private snapshots: %w", err)
	}

	userID := gameCtx.UserID()

	snap, ok := snapshots[userID]
	if !ok {
		return fmt.Errorf("no private snapshot for connecting player %s", userID)
	}

	missionResolver := BuildMissionResolver(p.missionController)

	msgs, err := converter.ConvertPrivateSnapshot(gameCtx, snap, missionResolver)
	if err != nil {
		return fmt.Errorf("failed to convert private snapshot for %s: %w", userID, err)
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
			return fmt.Errorf("failed to build %s message for %s: %w",
				item.msgType, userID, err)
		}

		p.writer.WriteMessage(gameCtx, msg)
	}

	return nil
}

func (p *GameStateBroadcaster) publishMoveHistory(gameCtx gamectx.GameContext) error {
	history, err := p.moveLogController.GetMoveLogs(gameCtx, p.historyConfig.Size)
	if err != nil {
		return fmt.Errorf("failed to get move history: %w", err)
	}

	msg, err := messaging.BuildMessage(messaging.MoveHistoryType, history)
	if err != nil {
		return fmt.Errorf("failed to build move history message: %w", err)
	}

	p.writer.WriteMessage(gameCtx, msg)

	return nil
}
