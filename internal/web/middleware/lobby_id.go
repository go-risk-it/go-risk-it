package middleware

import (
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type LobbyMiddleware struct{}

func NewLobbyMiddleware() *LobbyMiddleware {
	return &LobbyMiddleware{}
}

func (g *LobbyMiddleware) Wrap(routeToWrap *route.Route) *route.Route {
	if !strings.HasPrefix(routeToWrap.Pattern(), "/api/v1/lobbies/{id}") {
		return routeToWrap
	}

	return route.New(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		buildDomainContext[ctx.LobbyContext](
			routeToWrap,
			"lobby",
			ctx.WithLobbyID,
		),
	)
}
