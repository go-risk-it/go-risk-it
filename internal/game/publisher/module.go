package consumers

import (
	gameconfig "github.com/go-risk-it/go-risk-it/internal/game/config"
	"github.com/go-risk-it/go-risk-it/internal/game/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"go.uber.org/fx"
)

// Params groups the publisher's dependencies for fx injection.
type Params struct {
	fx.In

	Bus               bus.Bus
	Writer            Writer
	Presence          Presence
	Lifecycle         Lifecycle
	SnapshotService   snapshot.Service
	MissionController *MissionController
	MoveLogController *MoveLogController
	HistoryConfig     gameconfig.HistoryConfig
	Metrics           *metrics.InfraMetrics
}

func newGameStatePublisher(params Params) *GameStatePublisher {
	return NewGameStatePublisher(
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

func register(params Params, publisher *GameStatePublisher) {
	publisher.Register(params.Bus)
}

var Module = fx.Options(
	fx.Provide(
		NewMissionController,
		NewMoveLogController,
		newGameStatePublisher,
	),
	fx.Invoke(register),
)
