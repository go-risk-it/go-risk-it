package ctx

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type GameContext interface {
	UserContext
	Detachable
	GameID() int64
	WithBase(base context.Context) GameContext
}

type gameContext struct {
	UserContext

	gameID int64
}

var (
	_ GameContext = (*gameContext)(nil)
	_ LogEnricher = (*gameContext)(nil)
)

func (c *gameContext) GameID() int64 {
	return c.gameID
}

func (c *gameContext) SlogAttrs() []slog.Attr {
	return append(c.UserContext.SlogAttrs(), slog.Int64("gameID", c.gameID))
}

func (c *gameContext) DetachOnto(base context.Context) context.Context {
	return WithGameID(
		WithUserID(WithSpan(base, trace.SpanFromContext(base)), c.UserID()),
		c.gameID,
	)
}

func (c *gameContext) WithBase(base context.Context) GameContext {
	return &gameContext{
		UserContext: &userContext{
			TraceContext: &traceContext{
				Context: base,
				span:    trace.SpanFromContext(base),
			},
			userID: c.UserID(),
		},
		gameID: c.gameID,
	}
}

func WithGameID(ctx UserContext, gameID int64) GameContext {
	return &gameContext{
		UserContext: ctx,
		gameID:      gameID,
	}
}
