package game

import (
	"github.com/go-risk-it/go-risk-it/internal/game/consumers"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/rest"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	controller.Module,
	rest.Module,
	ws.Module,
	// Adapt ws.Manager to consumer-local interfaces via duck typing.
	fx.Provide(func(m ws.Manager) consumers.Writer { return m }),
	fx.Provide(func(m ws.Manager) consumers.Presence { return m }),
	fx.Provide(func(m ws.Manager) consumers.Lifecycle { return m }),
	consumers.Module,
)
