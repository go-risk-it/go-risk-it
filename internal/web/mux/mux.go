package mux

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewServeMux(
	routes []route.Route,
	authMiddleware middleware.AuthMiddleware,
	corsMiddleware middleware.CorsMiddleware,
	gameMiddleware middleware.GameMiddleware,
	lobbyMiddleware middleware.LobbyMiddleware,
	logMiddleware middleware.LogMiddleware,
	otelMiddleware middleware.OTelMiddleware,
	websocketAuthMiddleware middleware.WebsocketHeaderConversionMiddleware,
	log *zap.SugaredLogger,
) http.Handler {
	mux := http.NewServeMux()
	routeNames := make([]string, 0, len(routes))

	for _, route := range routes {
		mux.Handle(
			route.Pattern(),
			otelhttp.WithRouteTag(
				route.Pattern(),
				logMiddleware.Wrap(
					otelMiddleware.Wrap(
						corsMiddleware.Wrap(
							websocketAuthMiddleware.Wrap(
								authMiddleware.Wrap(
									lobbyMiddleware.Wrap(
										gameMiddleware.Wrap(
											route,
										),
									),
								),
							),
						),
					),
				),
			),
		)

		routeNames = append(routeNames, route.Pattern())
	}

	log.Infow("Registered routes", "routes", routeNames)

	handler := otelhttp.NewHandler(mux, "/")

	return handler
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewServeMux,
			fx.ParamTags(`group:"routes"`),
		),
	),
)
