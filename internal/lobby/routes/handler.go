package routes

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

// LobbyHandler handles requests with an enriched LobbyContext.
type LobbyHandler func(http.ResponseWriter, *http.Request, ctx.LobbyContext) error

// Lobby creates an authenticated route that extracts {id} into a LobbyContext.
func Lobby(pattern string, handler LobbyHandler) *route.Route {
	return route.New(
		pattern,
		true,
		route.WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
			lobbyCtx, err := BuildLobbyContext(request)
			if err != nil {
				return err
			}

			return handler(writer, request, lobbyCtx)
		}),
	)
}

// LobbyWS creates an authenticated WebSocket route with LobbyContext.
func LobbyWS(pattern string, handler LobbyHandler) *route.Route {
	return route.NewWS(
		pattern,
		true,
		route.WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
			lobbyCtx, err := BuildLobbyContext(request)
			if err != nil {
				return err
			}

			return handler(writer, request, lobbyCtx)
		}),
	)
}
