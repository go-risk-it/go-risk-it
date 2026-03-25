package testonly

import (
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewController,
		NewService,
		fx.Annotate(
			ProvideRoutes,
			fx.ResultTags(`group:"routes,flatten"`),
		),
	),
)

func ProvideRoutes(ctrl Controller) []*route.Route {
	return []*route.Route{
		NewResetHandler(ctrl),
		NewSetupNearWinHandler(ctrl),
	}
}
