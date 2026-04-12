package safego

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

// SafeOp runs action with a child span. On panic it records the error on the span
// and logs the recovered value and stack trace. This is a sequential wrapper
// (not a goroutine) — the bus already owns goroutine lifecycle.
//
// The action receives the span-enriched context and returns an error. SafeOp
// passes the error (or nil) to done(err) so the span records success/failure.
func SafeOp(
	parent context.Context,
	name string,
	action func(ctx context.Context) error,
) {
	ctx, done := observe.RawSpan(parent, "consumer."+name, attribute.String("handler", name))

	defer func() {
		if recovered := recover(); recovered != nil {
			done(fmt.Errorf("panic in %s: %v", name, recovered))

			slog.ErrorContext(ctx, "panic in consumer operation",
				"operation", name,
				"error", recovered,
				"stack", string(debug.Stack()),
			)
		}
	}()

	err := action(ctx)
	done(err)
}

// TypedSafeOp wraps SafeOp with typed context propagation. The action receives
// the domain-typed context directly — no type assertion needed at call sites.
//
// Rebaseable.Rebase contract guarantees the context type is preserved through
// the SafeOp span boundary.
//
//nolint:forcetypeassert // Rebaseable.Rebase contract guarantees type preservation
func TypedSafeOp[C ctx.Rebaseable](c C, name string, action func(C) error) {
	SafeOp(c, name, func(rawCtx context.Context) error {
		return action(rawCtx.(C))
	})
}
