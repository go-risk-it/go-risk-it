package publisher

import (
	"github.com/go-risk-it/go-risk-it/internal/events"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	Bus       events.Bus
	Publisher *LobbyStatePublisher
}

func register(params Params) {
	params.Publisher.Register(params.Bus)
}

var Module = fx.Options(
	fx.Provide(NewLobbyStatePublisher),
	fx.Invoke(register),
)
