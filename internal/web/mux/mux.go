package mux

import (
	"log/slog"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/fx"
)

func NewServeMux(
	routes []*route.Route,
	authMiddleware *middleware.AuthMiddleware,
	corsMiddleware *middleware.CorsMiddleware,
	logMiddleware *middleware.LogMiddleware,
	otelMiddleware *middleware.OTelMiddleware,
) http.Handler {
	mux := http.NewServeMux()
	routeNames := make([]string, 0, len(routes))

	for _, route := range routes {
		mux.Handle(
			route.Pattern(),
			logMiddleware.Wrap(
				otelMiddleware.Wrap(
					corsMiddleware.Wrap(
						authMiddleware.Wrap(
							route,
						),
					),
				),
			),
		)

		routeNames = append(routeNames, route.Pattern())
	}

	slog.Info("Registered routes", "routes", routeNames)

	return otelhttp.NewHandler(mux, "/")
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewServeMux,
			fx.ParamTags(`group:"routes"`),
		),
	),
)
