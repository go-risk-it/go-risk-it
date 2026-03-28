package publisher

import (
	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/events"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"go.uber.org/fx"
)

// Params groups the publisher's dependencies for fx injection.
type Params struct {
	fx.In

	Bus               events.Bus
	ConnectionManager ws.Manager
	SnapshotService   snapshot.Service
	MissionController *controller.MissionController
	MoveLogController *controller.MoveLogController
	HistoryConfig     config.HistoryConfig
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
	)
}

func register(params Params, publisher *GameStatePublisher) {
	publisher.Register(params.Bus)
}

var Module = fx.Options(
	fx.Provide(newGameStatePublisher),
	fx.Invoke(register),
)
