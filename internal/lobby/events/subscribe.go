package lobby

import (
	"context"
	"log/slog"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

// OnLobbyEvent registers a typed handler on the bus for lobby events of type E.
// It wraps eventbus.OnEvent[E] with a context assertion layer: the bus dispatches
// with context.Context, and this helper narrows it to ctx.LobbyContext before
// invoking the typed handler.
//
// If the dispatched context is not a LobbyContext (which should not happen in
// production when emitters use the correct context), the handler logs an error
// and returns as a no-op.
func OnLobbyEvent[E LobbyEvent](bus eventbus.Bus, handler func(ctx.LobbyContext, E)) {
	eventbus.OnEvent[E](bus, func(rawCtx context.Context, event E) {
		lobbyCtx, ok := rawCtx.(ctx.LobbyContext)
		if !ok {
			slog.ErrorContext(
				rawCtx,
				"OnLobbyEvent: context is not LobbyContext, skipping handler",
				slog.String("event_type", event.EventType()),
			)

			return
		}

		handler(lobbyCtx, event)
	})
}
