package ctx

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type LobbyContext interface {
	UserContext
	Detachable
	LobbyID() int64
}

type lobbyContext struct {
	UserContext

	lobbyID int64
}

var (
	_ LobbyContext = (*lobbyContext)(nil)
	_ LogEnricher  = (*lobbyContext)(nil)
)

func (c *lobbyContext) LobbyID() int64 {
	return c.lobbyID
}

func (c *lobbyContext) SlogAttrs() []slog.Attr {
	return append(c.UserContext.SlogAttrs(), slog.Int64("lobbyID", c.lobbyID))
}

func (c *lobbyContext) DetachOnto(base context.Context) context.Context {
	return WithLobbyID(
		WithUserID(WithSpan(base, trace.SpanFromContext(base)), c.UserID()),
		c.lobbyID,
	)
}

func WithLobbyID(ctx UserContext, lobbyID int64) LobbyContext {
	return &lobbyContext{
		UserContext: ctx,
		lobbyID:     lobbyID,
	}
}
