package middleware

import (
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/zap"
)

type GameMiddleware struct {
	log *zap.SugaredLogger
}

func NewGameMiddleware(log *zap.SugaredLogger) *GameMiddleware {
	return &GameMiddleware{log: log}
}

func (g *GameMiddleware) Wrap(routeToWrap route.Route) route.Route {
	if !strings.HasPrefix(routeToWrap.Pattern(), "/api/v1/games/{id}") {
		return routeToWrap
	}

	return route.NewRoute(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		buildDomainContext[ctx.GameContext](
			g.log,
			routeToWrap,
			"game",
			ctx.WithGameID,
		),
	)
}
