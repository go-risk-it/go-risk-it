package ctx

import "log/slog"

type UserContext interface {
	TraceContext
	LogEnricher
	UserID() string
}

type userContext struct {
	TraceContext

	userID string
}

var (
	_ UserContext = (*userContext)(nil)
	_ LogEnricher = (*userContext)(nil)
)

func (c *userContext) UserID() string {
	return c.userID
}

func (c *userContext) SlogAttrs() []slog.Attr {
	return []slog.Attr{slog.String("userID", c.userID)}
}

func WithUserID(ctx TraceContext, userID string) UserContext {
	return &userContext{
		TraceContext: ctx,
		userID:       userID,
	}
}
