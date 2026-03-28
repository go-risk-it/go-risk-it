package publisher

import (
	"github.com/go-risk-it/go-risk-it/internal/events"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"go.uber.org/fx"
)

// Params groups the publisher's dependencies for fx injection.
type Params struct {
	fx.In

	Bus               events.Bus
	ConnectionManager ws.Manager
	StateController   *controller.StateController
}

func newLobbyStatePublisher(params Params) *LobbyStatePublisher {
	return NewLobbyStatePublisher(
		params.ConnectionManager,
		params.StateController,
	)
}

func register(params Params, publisher *LobbyStatePublisher) {
	publisher.Register(params.Bus)
}

var Module = fx.Options(
	fx.Provide(newLobbyStatePublisher),
	fx.Invoke(register),
)
