package routes

import (
	"go.uber.org/fx"
)

// Module provides game REST controllers and route registrations.
var Module = fx.Options(
	fx.Provide(
		NewGameController,
		NewMoveController,
		NewAdvancementController,
		fx.Annotate(
			ProvideRoutes,
			fx.ResultTags(`group:"routes,flatten"`),
		),
		fx.Annotate(
			ProvideWSRoute,
			fx.ResultTags(`group:"routes"`),
		),
	),
)
