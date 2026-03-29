package slog

import (
	"context"
	"fmt"
	stdslog "log/slog"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

// ContextHandler is an slog.Handler that auto-extracts structured fields
// from the context chain via the ctx.LogEnricher interface. Any context type
// implementing SlogAttrs() is supported without concrete type switches.
// TraceID and SpanID are NOT extracted here — the otelslog bridge
// auto-extracts them from context.Context, so manual injection would
// create duplicates.
type ContextHandler struct {
	inner stdslog.Handler
	level stdslog.Level
}

var _ stdslog.Handler = (*ContextHandler)(nil)

func NewContextHandler(inner stdslog.Handler, level stdslog.Level) *ContextHandler {
	return &ContextHandler{inner: inner, level: level}
}

func (h *ContextHandler) Enabled(_ context.Context, level stdslog.Level) bool {
	return level >= h.level
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
	return &ContextHandler{inner: h.inner.WithAttrs(attrs), level: h.level}
}

func (h *ContextHandler) WithGroup(name string) stdslog.Handler {
	return &ContextHandler{inner: h.inner.WithGroup(name), level: h.level}
}

func extractContextAttrs(reqCtx context.Context) []stdslog.Attr {
	if enricher, ok := reqCtx.(ctx.LogEnricher); ok {
		return enricher.SlogAttrs()
	}

	return nil
}
