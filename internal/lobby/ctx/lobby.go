package ctx

import (
	"context"
	"log/slog"

	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"go.opentelemetry.io/otel/trace"
)

// LobbyContext enriches a [kernelctx.UserContext] with a lobby ID. It composes
// user identity and OTel tracing with lobby-scoped metadata so that every layer
// (logic, events, web) has typed access to the active lobby.
type LobbyContext interface {
	kernelctx.UserContext
	kernelctx.Rebaseable
	kernelctx.Scoped
	LobbyID() int64
}

type lobbyContext struct {
	kernelctx.UserContext

	lobbyID int64
}

var (
	_ LobbyContext          = (*lobbyContext)(nil)
	_ kernelctx.LogEnricher = (*lobbyContext)(nil)
	_ kernelctx.Rebaseable  = (*lobbyContext)(nil)
	_ kernelctx.Scoped      = (*lobbyContext)(nil)
)

func (c *lobbyContext) LobbyID() int64 {
	return c.lobbyID
}

func (c *lobbyContext) ScopeID() int64 {
	return c.lobbyID
}

func (c *lobbyContext) SlogAttrs() []slog.Attr {
	return append(c.UserContext.SlogAttrs(), slog.Int64("lobby_id", c.lobbyID))
}

func (c *lobbyContext) Rebase(base context.Context) context.Context {
	return WithLobbyID(
		kernelctx.WithUserID(kernelctx.WithSpan(base, trace.SpanFromContext(base)), c.UserID()),
		c.lobbyID,
	)
}

// WithLobbyID creates a LobbyContext from a UserContext and a lobby ID.
//

func WithLobbyID(ctx kernelctx.UserContext, lobbyID int64) LobbyContext {
	return &lobbyContext{
		UserContext: ctx,
		lobbyID:     lobbyID,
	}
}
