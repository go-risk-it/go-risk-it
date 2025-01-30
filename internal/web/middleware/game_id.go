package middleware

import (
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/zap"
)

type GameMiddleware interface {
	Middleware
}

type GameMiddlewareImpl struct {
	log *zap.SugaredLogger
}

var _ GameMiddleware = (*GameMiddlewareImpl)(nil)

func NewGameMiddleware(log *zap.SugaredLogger) GameMiddleware {
	return &GameMiddlewareImpl{log: log}
}

func (g *GameMiddlewareImpl) Wrap(routeToWrap route.Route) route.Route {
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
