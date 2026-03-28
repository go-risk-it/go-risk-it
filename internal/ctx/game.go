package ctx

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace/noop"
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

var _ GameContext = (*gameContext)(nil)

func (c *gameContext) GameID() int64 {
	return c.gameID
}

func (c *gameContext) Detach(timeout time.Duration) (context.Context, context.CancelFunc) {
	return DetachGameContextWithTimeout(c, timeout)
}

func (c *gameContext) WithBase(base context.Context) GameContext {
	return &gameContext{
		UserContext: &userContext{
			TraceContext: &traceContext{
				Context: base,
				span:    c.Span(),
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

// DetachGameContextWithTimeout creates a detached GameContext with a timeout.
// The returned cancel function must be called to release resources.
// This is useful for fire-and-forget goroutines that should be cancelled
// if they don't complete within the timeout.
func DetachGameContextWithTimeout(
	original GameContext,
	timeout time.Duration,
) (GameContext, context.CancelFunc) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)

	return detachGameContext(original, timeoutCtx), cancel
}

func detachGameContext(original GameContext, parent context.Context) GameContext {
	traceCtx := WithSpan(parent, noop.Span{})
	userCtx := &userContext{
		TraceContext: traceCtx,
		userID:       original.UserID(),
	}

	return &gameContext{
		UserContext: userCtx,
		gameID:      original.GameID(),
	}
}
