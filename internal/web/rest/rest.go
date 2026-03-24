package rest

import (
	"github.com/go-risk-it/go-risk-it/internal/web/rest/health"
	"go.uber.org/fx"
)

var Module = fx.Options(
	health.Module,
	fx.Provide(
		fx.Annotate(
			ProvideRoutes,
			fx.ParamTags(`name:"healthRoute"`),
			fx.ResultTags(`group:"routes,flatten"`),
		),
	),
)
