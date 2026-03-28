package ctx

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace/noop"
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

var _ LobbyContext = (*lobbyContext)(nil)

func (c *lobbyContext) LobbyID() int64 {
	return c.lobbyID
}

func (c *lobbyContext) Detach(timeout time.Duration) (context.Context, context.CancelFunc) {
	return DetachLobbyContextWithTimeout(c, timeout)
}

func WithLobbyID(ctx UserContext, lobbyID int64) LobbyContext {
	return &lobbyContext{
		UserContext: ctx,
		lobbyID:     lobbyID,
	}
}

// DetachLobbyContextWithTimeout creates a detached LobbyContext with a timeout.
// The returned cancel function must be called to release resources.
// This is useful for fire-and-forget goroutines that should be cancelled
// if they don't complete within the timeout.
func DetachLobbyContextWithTimeout(
	original LobbyContext,
	timeout time.Duration,
) (LobbyContext, context.CancelFunc) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)

	return detachLobbyContext(original, timeoutCtx), cancel
}

func detachLobbyContext(original LobbyContext, parent context.Context) LobbyContext {
	traceCtx := WithSpan(parent, noop.Span{})
	userCtx := &userContext{
		TraceContext: traceCtx,
		userID:       original.UserID(),
	}

	return &lobbyContext{
		UserContext: userCtx,
		lobbyID:     original.LobbyID(),
	}
}
