package nbio

import (
	"net/http"

	"github.com/lesismal/nbio/nbhttp"
)

func newNbioConfig(mux *http.ServeMux) nbhttp.Config {
	return nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{"localhost:8080"},
		MaxLoad:                 1000000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
	}
}
