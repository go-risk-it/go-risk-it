package game

import (
	consumers "github.com/go-risk-it/go-risk-it/internal/game/publisher"
	"github.com/go-risk-it/go-risk-it/internal/game/routes"
	"github.com/go-risk-it/go-risk-it/internal/game/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	routes.Module,
	ws.Module,
	// Adapt ws.Manager to consumer-local interfaces via duck typing.
	fx.Provide(func(m ws.Manager) consumers.Writer { return m }),
	fx.Provide(func(m ws.Manager) consumers.Presence { return m }),
	fx.Provide(func(m ws.Manager) consumers.Lifecycle { return m }),
	consumers.Module,
)
