package publisher

import (
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"go.uber.org/fx"
)

// Params groups the publisher's dependencies for fx injection.
type Params struct {
	fx.In

	Bus               bus.Bus
	ConnectionManager ws.Manager
	StateController   *controller.StateController
	Metrics           *metrics.Metrics
}

func newLobbyStatePublisher(params Params) *LobbyStatePublisher {
	return NewLobbyStatePublisher(
		params.ConnectionManager,
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
