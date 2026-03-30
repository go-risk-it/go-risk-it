package game

import (
	"context"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
)

// OnGameEvent registers a typed handler on the bus for game events of type E.
// It wraps eventbus.OnEvent[E] with a context assertion: if the dispatched context
// is not a ctx.GameContext, the handler logs an error and returns (no-op).
// This centralizes the context assertion that all game bus consumers need.
func OnGameEvent[E GameEvent](sub eventbus.Subscriber, handler func(ctx.GameContext, E)) {
	eventbus.OnEvent[E](sub, func(rawCtx context.Context, event E) {
		gameCtx, ok := rawCtx.(ctx.GameContext)
		if !ok {
			slog.ErrorContext(rawCtx, "OnGameEvent: context is not GameContext, skipping handler",
				slog.String("event_type", event.EventType()),
			)

			return
		}

		handler(gameCtx, event)
	})
}
