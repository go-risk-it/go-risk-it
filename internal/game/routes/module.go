package routes

import (
	"github.com/go-risk-it/go-risk-it/internal/game/commands"
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
	// Export GameController as commands.Handler for the kernel router.
	fx.Provide(func(gc *GameController) commands.Handler { return gc }),
)
