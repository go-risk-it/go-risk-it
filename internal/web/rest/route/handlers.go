package route

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
)

// PlainHandler handles requests with standard http types, returning errors.
type PlainHandler func(http.ResponseWriter, *http.Request) error

// GameHandler handles requests with an enriched GameContext.
type GameHandler func(http.ResponseWriter, *http.Request, ctx.GameContext) error

// LobbyHandler handles requests with an enriched LobbyContext.
type LobbyHandler func(http.ResponseWriter, *http.Request, ctx.LobbyContext) error
