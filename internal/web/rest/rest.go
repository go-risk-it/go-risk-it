package rest

import (
	"github.com/go-risk-it/go-risk-it/internal/web/rest/game"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/health"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/fx"
)

var Module = fx.Options(
	game.Module,
	health.Module,
	fx.Provide(
		route.AsRoute(NewWebSocketHandler),
	),
)
