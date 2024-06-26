package health

import (
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		route.AsRoute(New),
	),
)
