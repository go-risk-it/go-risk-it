package lobby

import (
	lobbydata "github.com/go-risk-it/go-risk-it/internal/lobby/internal/data"
	lobbylogic "github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic"
	weblobby "github.com/go-risk-it/go-risk-it/internal/lobby/web"
	"github.com/go-risk-it/go-risk-it/internal/lobby/web/routes"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	lobbylogic.Module,
	lobbydata.Module,
	routes.Module,
	ws.Module,
	// Adapt ws.Manager to consumer-local interfaces via duck typing.
	fx.Provide(func(m ws.Manager) weblobby.Writer { return m }),
	weblobby.Module,
)
