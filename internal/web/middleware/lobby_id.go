package middleware

import (
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/zap"
)

type LobbyMiddleware interface {
	Middleware
}

type LobbyMiddlewareImpl struct {
	log *zap.SugaredLogger
}

var _ LobbyMiddleware = (*LobbyMiddlewareImpl)(nil)

func NewLobbyMiddleware(log *zap.SugaredLogger) LobbyMiddleware {
	return &LobbyMiddlewareImpl{log: log}
}

func (g *LobbyMiddlewareImpl) Wrap(routeToWrap route.Route) route.Route {
	if !strings.HasPrefix(routeToWrap.Pattern(), "/api/v1/lobbies/{id}") {
		return routeToWrap
	}

	return route.NewRoute(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		buildDomainContext[ctx.LobbyContext](
			g.log,
			routeToWrap,
			"lobby",
			ctx.WithLobbyID,
		),
	)
}
