package controller

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewCreationController,
			fx.As(new(CreationController)),
		),
		fx.Annotate(
			NewManagementController,
			fx.As(new(ManagementController)),
		),
		fx.Annotate(
			NewStateController,
			fx.As(new(StateController)),
		),
		fx.Annotate(
			NewStartController,
			fx.As(new(StartController)),
		),
	),
)
