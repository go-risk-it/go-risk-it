package route

import (
	"net/http"
)

// Public creates an unauthenticated route with error handling.
func Public(pattern string, handler PlainHandler) *Route {
	return &Route{
		pattern:      pattern,
		requiresAuth: false,
		handler:      WrapErrors(handler),
	}
}

// Authed creates an authenticated route with error handling.
func Authed(pattern string, handler PlainHandler) *Route {
	return &Route{
		pattern:      pattern,
		requiresAuth: true,
		handler:      WrapErrors(handler),
	}
}

// Game creates an authenticated route that extracts {id} into a GameContext.
func Game(pattern string, handler GameHandler) *Route {
	return &Route{
		pattern:      pattern,
		requiresAuth: true,
		handler: WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
			gameCtx, err := BuildGameContext(request)
			if err != nil {
				return err
			}

			return handler(writer, request, gameCtx)
		}),
	}
}

// Lobby creates an authenticated route that extracts {id} into a LobbyContext.
func Lobby(pattern string, handler LobbyHandler) *Route {
	return &Route{
		pattern:      pattern,
		requiresAuth: true,
		handler: WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
			lobbyCtx, err := BuildLobbyContext(request)
			if err != nil {
				return err
			}

			return handler(writer, request, lobbyCtx)
		}),
	}
}

// GameWS creates an authenticated WebSocket route with GameContext.
func GameWS(pattern string, handler GameHandler) *Route {
	return &Route{
		pattern:      pattern,
		requiresAuth: true,
		isWebSocket:  true,
		handler: WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
			gameCtx, err := BuildGameContext(request)
			if err != nil {
				return err
			}

			return handler(writer, request, gameCtx)
		}),
	}
}

// LobbyWS creates an authenticated WebSocket route with LobbyContext.
func LobbyWS(pattern string, handler LobbyHandler) *Route {
	return &Route{
		pattern:      pattern,
		requiresAuth: true,
		isWebSocket:  true,
		handler: WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
			lobbyCtx, err := BuildLobbyContext(request)
			if err != nil {
				return err
			}

			return handler(writer, request, lobbyCtx)
		}),
	}
}
