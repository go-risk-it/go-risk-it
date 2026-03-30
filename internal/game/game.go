package game

import (
	gameconfig "github.com/go-risk-it/go-risk-it/internal/game/config"
	consumers "github.com/go-risk-it/go-risk-it/internal/game/consumers"
	"github.com/go-risk-it/go-risk-it/internal/game/headlines"
	"github.com/go-risk-it/go-risk-it/internal/game/routes"
	"github.com/go-risk-it/go-risk-it/internal/game/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	gameconfig.Module,
	headlines.Module,
	snapshot.Module,
	routes.Module,
	ws.Module,
	// Adapt ws.Manager to broadcaster-local interfaces via duck typing.
	fx.Provide(func(m ws.Manager) consumers.Writer { return m }),
	fx.Provide(func(m ws.Manager) consumers.Presence { return m }),
	fx.Provide(func(m ws.Manager) consumers.Lifecycle { return m }),
	// Adapt ws.Manager to the narrow Gateway interface for the route handler.
	fx.Provide(func(m ws.Manager) ws.Gateway { return m }),
	consumers.Module,
)
