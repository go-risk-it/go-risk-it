package nbio

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/lesismal/llib/std/crypto/tls"
	"github.com/lesismal/nbio/nbhttp"
	"go.uber.org/zap"
)

func newNbioConfig(
	mux *http.ServeMux,
	tlsConfig config.TLSConfig,
	log *zap.SugaredLogger,
) nbhttp.Config {
	if len(tlsConfig.Key) == 0 || len(tlsConfig.Cert) == 0 {
		log.Fatalf("TLS key and cert must be provided")
	}

	cert, err := tls.X509KeyPair(tlsConfig.Cert, tlsConfig.Key)
	if err != nil {
		log.Fatalf("failed to load TLS key and cert: %v", err)
	}

	log.Infow("TLS key and cert loaded")

	return nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{"0.0.0.0:8080"},
		AddrsTLS:                []string{"0.0.0.0:9443"},
		MaxLoad:                 1000000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
		TLSConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
		},
	}
}
