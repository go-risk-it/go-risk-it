package lobby

import (
	consumers "github.com/go-risk-it/go-risk-it/internal/lobby/consumers"
	"github.com/go-risk-it/go-risk-it/internal/lobby/routes"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	routes.Module,
	ws.Module,
	// Adapt ws.Manager to consumer-local interfaces via duck typing.
	fx.Provide(func(m ws.Manager) consumers.Writer { return m }),
	consumers.Module,
)
