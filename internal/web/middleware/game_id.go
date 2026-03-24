package middleware

import (
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type GameMiddleware struct{}

func NewGameMiddleware() *GameMiddleware {
	return &GameMiddleware{}
}

func (g *GameMiddleware) Wrap(routeToWrap *route.Route) *route.Route {
	if !strings.HasPrefix(routeToWrap.Pattern(), "/api/v1/games/{id}") {
		return routeToWrap
	}

	return route.New(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		buildDomainContext[ctx.GameContext](
			routeToWrap,
			"game",
			ctx.WithGameID,
		),
	)
}
