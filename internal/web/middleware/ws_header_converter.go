package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type WebsocketHeaderConversionMiddleware interface {
	Wrap(handler http.Handler) http.Handler
}

type WebsocketHeaderConversionMiddlewareImpl struct {
	log *zap.SugaredLogger
}

func NewWebsocketAuthMiddleware(log *zap.SugaredLogger) WebsocketHeaderConversionMiddleware {
	return &WebsocketHeaderConversionMiddlewareImpl{
		log: log,
	}
}

// Wrap extracts the token from the subprotocol and adds it to the HTTP Authorization header.
// Since Javascript's Websockets API sucks, we are forced to smuggle the token in a
// custom websocket protocol instead of simply using the HTTP Authorization header.
// The token is sent in the form of
//
//	"risk-it.websocket.auth.token, <token>" in the Sec-WebSocket-Protocol header.
//
// See: https://stackoverflow.com/questions/4361173/http-headers-in-websockets-client-api/77060459
func (m *WebsocketHeaderConversionMiddlewareImpl) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		m.log.Infow("wrapping request", "headers", request.Header)

		subprotocol := request.Header.Get("Sec-WebSocket-Protocol")
		if subprotocol != "" {
			token, err := extractToken(subprotocol)
			if err != nil {
				m.log.Errorw("unable to extract token from subprotocol", "error", err)

				return
			}

			request.Header.Set("Authorization", "Bearer "+token)
		}

		handler.ServeHTTP(writer, request)
	})
}

func extractToken(subprotocol string) (string, error) {
	if !strings.HasPrefix(subprotocol, "risk-it.websocket.auth.token, ") {
		return "", fmt.Errorf("invalid subprotocol")
	}

	return strings.TrimPrefix(subprotocol, "risk-it.websocket.auth.token, "), nil
}
