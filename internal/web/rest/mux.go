package rest

import (
	"net/http"

	"go.uber.org/zap"
)

func NewServeMux(routes []Route, log *zap.SugaredLogger) *http.ServeMux {
	mux := http.NewServeMux()

	for _, route := range routes {
		mux.Handle(route.Pattern(), route)
	}

	log.Infow("Created mux", "mux", mux, "routes", routes)

	return mux
}
