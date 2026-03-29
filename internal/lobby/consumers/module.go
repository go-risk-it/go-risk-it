package consumers

import (
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"go.uber.org/fx"
)

// Params groups the publisher's dependencies for fx injection.
type Params struct {
	fx.In

	Bus             bus.Bus
	Writer          Writer
	StateController *controller.StateController
	Metrics         *metrics.InfraMetrics
}

func newLobbyStatePublisher(params Params) *LobbyStatePublisher {
	return NewLobbyStatePublisher(
		params.Writer,
		params.StateController,
		params.Metrics,
	)
}

func register(params Params, publisher *LobbyStatePublisher) {
	publisher.Register(params.Bus)
}

var Module = fx.Options(
	fx.Provide(newLobbyStatePublisher),
	fx.Invoke(register),
)
