package publisher

import (
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"go.uber.org/fx"
)

// Params groups the publisher's dependencies for fx injection.
type Params struct {
	fx.In

	Bus               bus.Bus
	ConnectionManager ws.Manager
	SnapshotService   snapshot.Service
	MissionController *controller.MissionController
	MoveLogController *controller.MoveLogController
	HistoryConfig     config.HistoryConfig
	Metrics           *metrics.Metrics
}

func newGameStatePublisher(params Params) *GameStatePublisher {
	return NewGameStatePublisher(
		params.ConnectionManager,
		params.ConnectionManager,
		params.ConnectionManager,
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
	fx.Provide(newGameStatePublisher),
	fx.Invoke(register),
)
