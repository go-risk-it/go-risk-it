package rest

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"go.uber.org/zap"
)

func NewServeMux(
	routes []Route,
	authMiddleware middleware.AuthMiddleware,
	log *zap.SugaredLogger,
) *http.ServeMux {
	mux := http.NewServeMux()
	routeNames := make([]string, 0, len(routes))

	for _, route := range routes {
		mux.Handle(route.Pattern(), authMiddleware.Wrap(route))

		routeNames = append(routeNames, route.Pattern())
	}

	log.Infow("Registered routes", "routes", routeNames)

	return mux
}
