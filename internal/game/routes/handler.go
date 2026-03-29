package routes

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

// GameHandler handles requests with an enriched GameContext.
type GameHandler func(http.ResponseWriter, *http.Request, ctx.GameContext) error

// Game creates an authenticated route that extracts {id} into a GameContext.
func Game(pattern string, handler GameHandler) *route.Route {
	return route.New(
		pattern,
		true,
		route.WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
			gameCtx, err := BuildGameContext(request)
			if err != nil {
				return err
			}

			return handler(writer, request, gameCtx)
		}),
	)
}

// GameWS creates an authenticated WebSocket route with GameContext.
func GameWS(pattern string, handler GameHandler) *route.Route {
	return route.NewWS(
		pattern,
		true,
		route.WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
			gameCtx, err := BuildGameContext(request)
			if err != nil {
				return err
			}

			return handler(writer, request, gameCtx)
		}),
	)
}
