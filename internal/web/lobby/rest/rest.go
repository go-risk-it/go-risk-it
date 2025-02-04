package rest

import (
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		route.AsRoute(NewCreationHandler),
		route.AsRoute(NewJoinHandler),
		route.AsRoute(NewLobbiesHandler),
	),
)
