package ctx

import "context"

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
	logCtx := WithLog(context.Background(), original.Log())
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
