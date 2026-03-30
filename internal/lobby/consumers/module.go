package consumers

import (
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"go.uber.org/fx"
)

// Params groups the broadcaster's dependencies for fx injection.
type Params struct {
	fx.In

	Bus             bus.Subscriber
	Writer          Writer
	StateController *StateController
	Metrics         *metrics.InfraMetrics
}

func newLobbyStateBroadcaster(params Params) *LobbyStateBroadcaster {
	return NewLobbyStateBroadcaster(
		params.Writer,
		params.StateController,
		params.Metrics,
	)
}

func register(params Params, broadcaster *LobbyStateBroadcaster) {
	broadcaster.Register(params.Bus)
}

var Module = fx.Options(
	fx.Provide(
		NewStateController,
		newLobbyStateBroadcaster,
	),
	fx.Invoke(register),
)
