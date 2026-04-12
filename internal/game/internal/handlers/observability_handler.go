package handlers

import (
	"log/slog"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"go.uber.org/fx"
)

// observabilityHandler logs structured information about MoveCompleted events for
// tracing and debugging. Metrics recording (ActiveGames, GameDuration, summary
// histograms) is handled by the existing gameSummaryRecorder in the metrics package.
type observabilityHandler struct{}

// ObservabilityHandlerParams holds the dependencies for the observability handler.
type ObservabilityHandlerParams struct {
	fx.In

	Sub eventbus.Subscriber
}

// RegisterObservabilityHandler subscribes the observability handler to MoveCompleted events.
func RegisterObservabilityHandler(params ObservabilityHandlerParams) {
	h := &observabilityHandler{}

	gameevt.OnGameEvent[*gameevt.MoveCompleted](params.Sub, h.handleMoveCompleted)
}

func (h *observabilityHandler) handleMoveCompleted(
	gameCtx gamectx.GameContext,
	event *gameevt.MoveCompleted,
) {
	slog.DebugContext(gameCtx, "move completed",
		slog.Int64("gameId", event.GameID()),
		slog.String("actionType", string(event.ActionType)),
		slog.Int64("turn", event.Turn),
		slog.String("targetPhase", string(event.TargetPhase)),
		slog.Bool("gameOver", event.GameOver),
	)
}
