package mux

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewServeMux(
	routes []route.Route,
	authMiddleware middleware.AuthMiddleware,
	corsMiddleware middleware.CorsMiddleware,
	gameMiddleware middleware.GameMiddleware,
	logMiddleware middleware.LogMiddleware,
	websocketAuthMiddleware middleware.WebsocketHeaderConversionMiddleware,
	log *zap.SugaredLogger,
) *http.ServeMux {
	mux := http.NewServeMux()
	routeNames := make([]string, 0, len(routes))

	for _, route := range routes {
		mux.Handle(
			route.Pattern(),
			logMiddleware.Wrap(
				websocketAuthMiddleware.Wrap(
					authMiddleware.Wrap(
						corsMiddleware.Wrap(
							gameMiddleware.Wrap(
								route,
							),
						),
					),
				),
			),
		)

		routeNames = append(routeNames, route.Pattern())
	}

	log.Infow("Registered routes", "routes", routeNames)

	return mux
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewServeMux,
			fx.ParamTags(`group:"routes"`),
		),
	),
)
