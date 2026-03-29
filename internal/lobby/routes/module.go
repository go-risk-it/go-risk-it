package routes

import "go.uber.org/fx"

// Module provides lobby REST controllers and route registrations.
var Module = fx.Options(
	fx.Provide(
		NewCreationController,
		NewManagementController,
		NewStartController,
		fx.Annotate(
			ProvideRoutes,
			fx.ResultTags(`group:"routes,flatten"`),
		),
	),
)
