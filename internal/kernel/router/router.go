package router

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/commands"
)

// Router dispatches cross-module commands to the appropriate handler.
// It uses a type-switch to route commands, keeping modules decoupled.
type Router struct {
	gameHandler commands.Handler
}

// NewRouter creates a Router wired to the game module's command handler.
func NewRouter(gameHandler commands.Handler) *Router {
	return &Router{gameHandler: gameHandler}
}

// Route dispatches a command to the appropriate handler based on its type.
// Returns the handler's result or an error for unknown command types.
func (r *Router) Route(ctx context.Context, cmd any) (any, error) {
	switch c := cmd.(type) {
	case commands.CreateGame:
		return r.gameHandler.HandleCreateGame(ctx, c)
	default:
		return nil, fmt.Errorf("unknown command type: %T", cmd)
	}
}
