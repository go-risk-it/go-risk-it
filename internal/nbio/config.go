package nbio

import (
	"net/http"

	"github.com/lesismal/nbio/nbhttp"
	"go.uber.org/zap"
)

func NewNbioConfig(mux *http.ServeMux, log *zap.SugaredLogger) *nbhttp.Config {
	log.Infow("Using mux: ", "mux", mux)
	return &nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{"localhost:8080"},
		MaxLoad:                 1000000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
	}
}
