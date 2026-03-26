package signals

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/safego"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/converter"
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"go.uber.org/fx"
)

// PlayerConnectedHandlerParams is the dedicated params struct for
// HandlePlayerConnected, using the snapshot+converter path.
type PlayerConnectedHandlerParams struct {
	fx.In

	Signal            signals.PlayerConnectedSignal
	SnapshotService   snapshot.Service
	MissionController *controller.MissionController
	ConnectionManager ws.Manager
	MoveLogFetcher    fetcher.MoveLogFetcher
}

func HandlePlayerConnected(
	params PlayerConnectedHandlerParams,
) {
	params.Signal.AddListener(func(context context.Context, data signals.PlayerConnectedData) {
		gameContext, ok := context.(ctx.GameContext)
		if !ok {
			slog.ErrorContext(context, "context is not game context")

			return
		}

		slog.InfoContext(
			gameContext, "handling player connected",
		)

		safego.Go(gameContext, func() {
			publishPublicStateToPlayer(gameContext, params)
		})

		safego.Go(gameContext, func() {
			publishPrivateStateToPlayer(gameContext, params)
		})

		safego.Go(gameContext, func() {
			fetchStateAndPublish(
				gameContext,
				params.MoveLogFetcher.FetchState,
				params.ConnectionManager.WriteMessage,
			)
		})
	})
}

// publishPublicStateToPlayer fetches the public snapshot, converts it into WS
// messages, and writes each message to the connecting player (not broadcast).
func publishPublicStateToPlayer(
	gameContext ctx.GameContext,
	params PlayerConnectedHandlerParams,
) {
	detached, cancel := ctx.DetachGameContextWithTimeout(gameContext, fetchTimeout)
	defer cancel()

	snap, err := params.SnapshotService.GetPublicSnapshot(detached)
	if err != nil {
		slog.ErrorContext(detached, "failed to get public snapshot", "error", err)

		return
	}

	connectedPlayers := params.ConnectionManager.GetConnectedPlayers(detached)

	msgs, err := converter.ConvertPublicSnapshot(snap, connectedPlayers)
	if err != nil {
		slog.ErrorContext(detached, "failed to convert public snapshot", "error", err)

		return
	}

	for _, msg := range []json.RawMessage{
		msgs.GameState,
		msgs.BoardState,
		msgs.PlayerState,
	} {
		params.ConnectionManager.WriteMessage(detached, msg)
	}
}

// publishPrivateStateToPlayer fetches per-player private snapshots, extracts the
// connecting player's snapshot, converts it into WS messages, and writes them to
// the connecting player's connection.
func publishPrivateStateToPlayer(
	gameContext ctx.GameContext,
	params PlayerConnectedHandlerParams,
) {
	detached, cancel := ctx.DetachGameContextWithTimeout(gameContext, fetchTimeout)
	defer cancel()

	snapshots, err := params.SnapshotService.GetPrivateSnapshotsByUser(detached)
	if err != nil {
		slog.ErrorContext(detached, "failed to get private snapshots", "error", err)

		return
	}

	userID := detached.UserID()

	snap, ok := snapshots[userID]
	if !ok {
		slog.ErrorContext(detached, "no private snapshot for connecting player", "userID", userID)

		return
	}

	missionResolver := BuildMissionResolver(params.MissionController)

	msgs, err := converter.ConvertPrivateSnapshot(detached, snap, missionResolver)
	if err != nil {
		slog.ErrorContext(
			detached,
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
		params.ConnectionManager.WriteMessage(detached, msg)
	}
}
