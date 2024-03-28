package nbio

import (
	"net/http"

	"github.com/lesismal/nbio/nbhttp"
)

func newNbioConfig(mux *http.ServeMux) nbhttp.Config {
	return nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{"0.0.0.0:8080"},
		MaxLoad:                 1000000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
	}
}
