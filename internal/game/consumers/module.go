package consumers

import (
	gameconfig "github.com/go-risk-it/go-risk-it/internal/game/config"
	"github.com/go-risk-it/go-risk-it/internal/game/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"go.uber.org/fx"
)

// Params groups the broadcaster's dependencies for fx injection.
type Params struct {
	fx.In

	Bus               bus.Subscriber
	Writer            Writer
	Presence          Presence
	Lifecycle         Lifecycle
	SnapshotService   snapshot.Service
	MissionController *MissionController
	MoveLogController *MoveLogController
	HistoryConfig     gameconfig.HistoryConfig
	Metrics           *metrics.InfraMetrics
}

func newGameStateBroadcaster(params Params) *GameStateBroadcaster {
	return NewGameStateBroadcaster(
		params.Writer,
		params.Presence,
		params.Lifecycle,
		params.SnapshotService,
		params.MissionController,
		params.MoveLogController,
		params.HistoryConfig,
		params.Metrics,
	)
}

func register(params Params, broadcaster *GameStateBroadcaster) {
	broadcaster.Register(params.Bus)
}

var Module = fx.Options(
	fx.Provide(
		NewMissionController,
		NewMoveLogController,
		newGameStateBroadcaster,
	),
	fx.Invoke(register),
)
