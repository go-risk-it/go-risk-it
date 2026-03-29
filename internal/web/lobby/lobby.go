package lobby

import (
	"github.com/go-risk-it/go-risk-it/internal/lobby/consumers"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/rest"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	controller.Module,
	rest.Module,
	ws.Module,
	// Adapt ws.Manager to consumer-local interfaces via duck typing.
	fx.Provide(func(m ws.Manager) consumers.Writer { return m }),
	consumers.Module,
)
