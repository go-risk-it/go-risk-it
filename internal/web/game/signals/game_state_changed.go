package signals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
	"github.com/go-risk-it/go-risk-it/internal/safego"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/converter"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

func HandleGameStateChanged(
	params GameStateChangedHandlerParams,
) {
	params.Signal.AddListener(func(context context.Context, data signals.GameStateChangedData) {
		gameContext, ok := context.(ctx.GameContext)
		if !ok {
			slog.ErrorContext(context, "context is not game context")

			return
		}

		slog.InfoContext(
			gameContext,
			"handling game state changed",
			"fromPhase", data.FromPhase,
			"toPhase", data.ToPhase,
		)

		safego.Go(gameContext, func() {
			publishPublicState(gameContext, params)
		})

		safego.Go(gameContext, func() {
			publishPrivateStates(gameContext, params)
		})
	})
}

// publishPublicState fetches the public snapshot, converts it into WS
// messages, and broadcasts each message to all connected players.
func publishPublicState(
	gameContext ctx.GameContext,
	params GameStateChangedHandlerParams,
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
		params.ConnectionManager.Broadcast(detached, msg)
	}

	// After broadcasting the final game state, untrack the game from the WS
	// manager so completed games don't leak in the gameConnections map.
	// Safe to call here because getPlayerConnections (used by Broadcast/Write)
	// is read-only and no-ops when the game is already removed.
	if snap.Game.WinnerUserID.Valid {
		slog.InfoContext(detached, "game completed, removing from connection manager",
			"winner", snap.Game.WinnerUserID.String,
		)

		params.ConnectionManager.RemoveGame(detached)
	}
}

// publishPrivateStates fetches per-player private snapshots, converts
// each into WS messages (cards + mission), and writes them to the
// corresponding player's connection.
func publishPrivateStates(
	gameContext ctx.GameContext,
	params GameStateChangedHandlerParams,
) {
	detached, cancel := ctx.DetachGameContextWithTimeout(gameContext, fetchTimeout)
	defer cancel()

	snapshots, err := params.SnapshotService.GetPrivateSnapshotsByUser(detached)
	if err != nil {
		slog.ErrorContext(detached, "failed to get private snapshots", "error", err)

		return
	}

	missionResolver := BuildMissionResolver(params.MissionController)

	for userID, snap := range snapshots {
		msgs, err := converter.ConvertPrivateSnapshot(detached, snap, missionResolver)
		if err != nil {
			slog.ErrorContext(
				detached,
				"failed to convert private snapshot",
				"userID", userID,
				"error", err,
			)

			continue
		}

		playerCtx := ctx.WithGameID(
			ctx.WithUserID(ctx.WithSpan(detached, detached.Span()), userID),
			detached.GameID(),
		)

		for _, msg := range []json.RawMessage{
			msgs.CardState,
			msgs.MissionState,
		} {
			params.ConnectionManager.WriteMessage(playerCtx, msg)
		}
	}
}

// BuildMissionResolver creates a MissionResolver closure that dispatches
// to the correct MissionController method based on mission type, wrapping
// the typed result into a json.RawMessage envelope.
func BuildMissionResolver(
	missionController *controller.MissionController,
) converter.MissionResolver {
	return func(
		c context.Context,
		missionType sqlc.GameMissionType,
		missionID int64,
	) (json.RawMessage, error) {
		gameCtx, ok := c.(ctx.GameContext)
		if !ok {
			return nil, errors.New("mission resolver requires GameContext")
		}

		return resolveMission(gameCtx, missionController, missionType, missionID)
	}
}

// resolveMission dispatches to the correct MissionController method and
// wraps the result in a message envelope. Factored out of the closure to
// keep BuildMissionResolver within the cyclomatic complexity threshold.
func resolveMission(
	gameCtx ctx.GameContext,
	missionCtrl *controller.MissionController,
	missionType sqlc.GameMissionType,
	missionID int64,
) (json.RawMessage, error) {
	switch missionType {
	case sqlc.GameMissionTypeTWOCONTINENTS:
		return fetchAndBuildMission(missionCtrl.GetTwoContinentsMission, gameCtx, missionID)
	case sqlc.GameMissionTypeTWOCONTINENTSPLUSONE:
		return fetchAndBuildMission(missionCtrl.GetTwoContinentsPlusOneMission, gameCtx, missionID)
	case sqlc.GameMissionTypeELIMINATEPLAYER:
		return fetchAndBuildMission(missionCtrl.GetEliminatePlayerMission, gameCtx, missionID)
	case sqlc.GameMissionTypeEIGHTEENTERRITORIESTWOTROOPS:
		return fetchAndBuildMission(
			missionCtrl.GetEighteenTerritoriesTwoTroopsMission,
			gameCtx,
			missionID,
		)
	case sqlc.GameMissionTypeTWENTYFOURTERRITORIES:
		return fetchAndBuildMission(
			missionCtrl.GetTwentyFourTerritoriesMission,
			gameCtx,
			missionID,
		)
	default:
		return nil, fmt.Errorf("unknown mission type: %s", missionType)
	}
}

// fetchAndBuildMission is a generic helper that calls a typed mission
// controller method and wraps the result in a MissionState message envelope.
func fetchAndBuildMission[T any](
	fetch func(ctx.GameContext, int64) (T, error),
	gameCtx ctx.GameContext,
	missionID int64,
) (json.RawMessage, error) {
	state, err := fetch(gameCtx, missionID)
	if err != nil {
		return nil, fmt.Errorf("fetching mission: %w", err)
	}

	return message.BuildMessage(message.MissionState, state)
}
