package ctx

import (
	"context"
	"time"
)

type GameContext interface {
	UserContext
	GameID() int64
}

type gameContext struct {
	UserContext

	gameID int64
}

var _ GameContext = (*gameContext)(nil)

func (c *gameContext) GameID() int64 {
	return c.gameID
}

func WithGameID(ctx UserContext, gameID int64) GameContext {
	ctx.SetLog(ctx.Log().With("gameID", gameID))

	return &gameContext{
		UserContext: ctx,
		gameID:      gameID,
	}
}

// DetachGameContext creates a new GameContext rooted at context.Background(),
// preserving the logger, user ID, and game ID from the original context.
// This is useful for fire-and-forget goroutines that should not be cancelled
// when the originating HTTP request completes.
func DetachGameContext(original GameContext) GameContext {
	return detachGameContext(original, context.Background())
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
	logCtx := WithLog(parent, original.Log())
	userCtx := &userContext{
		TraceContext: &traceContext{
			LogContext: logCtx,
			span:       original.Span(),
		},
		userID: original.UserID(),
	}

	return &gameContext{
		UserContext: userCtx,
		gameID:      original.GameID(),
	}
}
