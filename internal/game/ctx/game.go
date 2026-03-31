package ctx

import (
	"context"
	"log/slog"

	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"go.opentelemetry.io/otel/trace"
)

// GameContext enriches a [kernelctx.UserContext] with a game ID. It composes
// user identity and OTel tracing with game-scoped metadata so that every layer
// (logic, events, web) has typed access to the active game.
type GameContext interface {
	kernelctx.UserContext
	kernelctx.Detachable
	GameID() int64
	WithBase(base context.Context) GameContext
}

type gameContext struct {
	kernelctx.UserContext

	gameID int64
}

var (
	_ GameContext           = (*gameContext)(nil)
	_ kernelctx.LogEnricher = (*gameContext)(nil)
)

func (c *gameContext) GameID() int64 {
	return c.gameID
}

func (c *gameContext) SlogAttrs() []slog.Attr {
	return append(c.UserContext.SlogAttrs(), slog.Int64("game_id", c.gameID))
}

func (c *gameContext) DetachOnto(base context.Context) context.Context {
	return WithGameID(
		kernelctx.WithUserID(kernelctx.WithSpan(base, trace.SpanFromContext(base)), c.UserID()),
		c.gameID,
	)
}

func (c *gameContext) WithBase(base context.Context) GameContext {
	return &gameContext{
		UserContext: kernelctx.WithUserID(
			kernelctx.WithSpan(base, trace.SpanFromContext(base)),
			c.UserID(),
		),
		gameID: c.gameID,
	}
}

// WithGameID creates a GameContext from a UserContext and a game ID.
//

func WithGameID(ctx kernelctx.UserContext, gameID int64) GameContext {
	return &gameContext{
		UserContext: ctx,
		gameID:      gameID,
	}
}
