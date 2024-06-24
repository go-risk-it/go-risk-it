package mux

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewServeMux(
	routes []rest.Route,
	authMiddleware middleware.AuthMiddleware,
	gameMiddleware middleware.GameMiddleware,
	websocketAuthMiddleware middleware.WebsocketHeaderConversionMiddleware,
	log *zap.SugaredLogger,
) *http.ServeMux {
	mux := http.NewServeMux()
	routeNames := make([]string, 0, len(routes))

	for _, route := range routes {
		mux.Handle(route.Pattern(),
			gameMiddleware.Wrap(
				websocketAuthMiddleware.Wrap(
					authMiddleware.Wrap(route))))

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
