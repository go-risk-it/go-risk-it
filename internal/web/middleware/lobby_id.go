package middleware

import (
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/zap"
)

type LobbyMiddleware struct {
	log *zap.SugaredLogger
}

func NewLobbyMiddleware(log *zap.SugaredLogger) *LobbyMiddleware {
	return &LobbyMiddleware{log: log}
}

func (g *LobbyMiddleware) Wrap(routeToWrap route.Route) route.Route {
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
