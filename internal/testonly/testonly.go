package testonly

import (
	"github.com/go-risk-it/go-risk-it/internal/web/rest"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewController,
			fx.As(new(Controller)),
		),
		fx.Annotate(
			NewService,
			fx.As(new(Service)),
		),
		rest.AsRoute(NewResetHandler),
	),
)
