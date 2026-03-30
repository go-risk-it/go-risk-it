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
	return route.Domain(pattern, BuildGameContext, handler)
}

// GameWS creates an authenticated WebSocket route with GameContext.
func GameWS(pattern string, handler GameHandler) *route.Route {
	return route.DomainWS(pattern, BuildGameContext, handler)
}
