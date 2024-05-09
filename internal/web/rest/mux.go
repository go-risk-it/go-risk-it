package rest

import (
	"net/http"

	"go.uber.org/zap"
)

func NewServeMux(routes []Route, log *zap.SugaredLogger) *http.ServeMux {
	mux := http.NewServeMux()
	routeNames := make([]string, 0, len(routes))

	for _, route := range routes {
		mux.Handle(route.Pattern(), route)

		routeNames = append(routeNames, route.Pattern())
	}

	log.Infow("Created mux", "mux", mux, "routes", routeNames)

	return mux
}
