package ctx

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type UserContext interface {
	TraceContext
	LogEnricher
	Rebaseable
	UserID() string
}

type userContext struct {
	TraceContext

	userID string
}

var (
	_ UserContext = (*userContext)(nil)
	_ LogEnricher = (*userContext)(nil)
	_ Rebaseable  = (*userContext)(nil)
)

func (c *userContext) UserID() string {
	return c.userID
}

func (c *userContext) SlogAttrs() []slog.Attr {
	return []slog.Attr{slog.String("user_id", c.userID)}
}

func (c *userContext) Rebase(base context.Context) context.Context {
	return WithUserID(WithSpan(base, trace.SpanFromContext(base)), c.userID)
}

func WithUserID(ctx TraceContext, userID string) UserContext {
	return &userContext{
		TraceContext: ctx,
		userID:       userID,
	}
}
