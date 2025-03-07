package rest

import (
	"github.com/go-risk-it/go-risk-it/internal/web/game/rest/move"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/fx"
)

var Module = fx.Options(
	move.Module,
	fx.Provide(
		route.AsRoute(NewAdvancementHandler),
		route.AsRoute(NewCreationHandler),
		route.AsRoute(NewManagementHandler),
	),
)
