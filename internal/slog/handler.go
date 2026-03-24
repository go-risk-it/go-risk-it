package slog

import (
	"context"
	"fmt"
	stdslog "log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
)

// ContextHandler is an slog.Handler that auto-extracts structured fields
// from the context chain (TraceContext, UserContext, GameContext, LobbyContext).
// This eliminates the need for SetLog() calls that manually enrich the logger
// at each context layer.
type ContextHandler struct {
	inner stdslog.Handler
}

var _ stdslog.Handler = (*ContextHandler)(nil)

func NewContextHandler(inner stdslog.Handler) *ContextHandler {
	return &ContextHandler{inner: inner}
}

func (h *ContextHandler) Enabled(reqCtx context.Context, level stdslog.Level) bool {
	return h.inner.Enabled(reqCtx, level)
}

func (h *ContextHandler) Handle(reqCtx context.Context, record stdslog.Record) error {
	attrs := extractContextAttrs(reqCtx)
	if len(attrs) > 0 {
		record.AddAttrs(attrs...)
	}

	if err := h.inner.Handle(reqCtx, record); err != nil {
		return fmt.Errorf("inner handler failed: %w", err)
	}

	return nil
}

func (h *ContextHandler) WithAttrs(attrs []stdslog.Attr) stdslog.Handler {
	return &ContextHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) stdslog.Handler {
	return &ContextHandler{inner: h.inner.WithGroup(name)}
}

func extractContextAttrs(reqCtx context.Context) []stdslog.Attr {
	var attrs []stdslog.Attr

	// Check for TraceContext (has Span) — must come before more specific checks
	// since GameContext and LobbyContext embed TraceContext.
	if traceCtx, ok := reqCtx.(ctx.TraceContext); ok {
		spanCtx := traceCtx.Span().SpanContext()
		if spanCtx.HasTraceID() {
			attrs = append(attrs, stdslog.String("traceID", spanCtx.TraceID().String()))
		}

		if spanCtx.HasSpanID() {
			attrs = append(attrs, stdslog.String("spanID", spanCtx.SpanID().String()))
		}
	}

	// Check for UserContext (has UserID)
	if userCtx, ok := reqCtx.(ctx.UserContext); ok {
		attrs = append(attrs, stdslog.String("userID", userCtx.UserID()))
	}

	// Check for GameContext (has GameID)
	if gameCtx, ok := reqCtx.(ctx.GameContext); ok {
		attrs = append(attrs, stdslog.Int64("gameID", gameCtx.GameID()))
	}

	// Check for LobbyContext (has LobbyID)
	if lobbyCtx, ok := reqCtx.(ctx.LobbyContext); ok {
		attrs = append(attrs, stdslog.Int64("lobbyID", lobbyCtx.LobbyID()))
	}

	return attrs
}
