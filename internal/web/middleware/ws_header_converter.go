package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/zap"
)

type WebsocketHeaderConversionMiddleware struct {
	log *zap.SugaredLogger
}

func NewWebsocketAuthMiddleware(log *zap.SugaredLogger) *WebsocketHeaderConversionMiddleware {
	return &WebsocketHeaderConversionMiddleware{
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
func (m *WebsocketHeaderConversionMiddleware) Wrap(routeToWrap *route.Route) *route.Route {
	return route.New(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			subprotocol := request.Header.Get("Sec-WebSocket-Protocol")
			if subprotocol != "" {
				token, err := extractToken(subprotocol)
				if err != nil {
					m.log.Errorw("unable to extract token from subprotocol", "error", err)

					return
				}

				request.Header.Set("Authorization", "Bearer "+token)
			}

			routeToWrap.ServeHTTP(writer, request)
		}))
}

func extractToken(subprotocol string) (string, error) {
	if !strings.HasPrefix(subprotocol, "risk-it.websocket.auth.token, ") {
		return "", errors.New("invalid subprotocol")
	}

	return strings.TrimPrefix(subprotocol, "risk-it.websocket.auth.token, "), nil
}
