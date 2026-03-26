package signals

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	logicsignals "github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/safego"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"go.uber.org/fx"
)

const fetchTimeout = 10 * time.Second

var Module = fx.Options(
	fx.Invoke(
		HandleGameStateChanged,
		HandleMovePerformed,
		HandlePlayerConnected,
	),
)

// GameStateChangedHandlerParams is the dedicated params struct for
// HandleGameStateChanged, using the snapshot+converter path instead of
// scattered fetchers.
type GameStateChangedHandlerParams struct {
	fx.In

	Signal            logicsignals.GameStateChangedSignal
	SnapshotService   snapshot.Service
	MissionController *controller.MissionController
	ConnectionManager ws.Manager
}

func fetchStateAndPublish(
	gameCtx ctx.GameContext,
	fetcher func(ctx.GameContext, chan json.RawMessage),
	publisher func(ctx.GameContext, json.RawMessage),
) {
	detached, cancel := ctx.DetachGameContextWithTimeout(gameCtx, fetchTimeout)
	defer cancel()

	channel := make(chan json.RawMessage, 1)
	safego.Go(detached, func() { fetcher(detached, channel) })

	select {
	case msg := <-channel:
		publisher(detached, msg)
	case <-detached.Done():
		slog.ErrorContext(detached, "timeout while fetching state", "timeout", fetchTimeout)
	}
}
