package observe

import (
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"go.opentelemetry.io/otel/attribute"
)

// Span wraps a typed-context function with automatic span lifecycle. The context
// type C is preserved through auto-rebase — fn receives the same typed context
// as parent. This is the primary span creation function for business logic.
//
// Eliminates named returns, defer-done discipline, and the risk of discarding
// the span-enriched context. For error-only functions, use [SpanErr]. For
// infrastructure callers on plain context.Context, use [RawSpan].
//
//	func (s *service) GetState(ctx gamectx.GameContext) (*Game, error) {
//	    return observe.Span(ctx, "game.get_state",
//	        func(ctx gamectx.GameContext) (*Game, error) {
//	            // ctx carries the child span — compile-time safe
//	        },
//	    )
//	}
//
//nolint:forcetypeassert // Rebaseable.Rebase contract guarantees type preservation
func Span[C ctx.Rebaseable, T any](
	parent C,
	name string,
	fn func(C) (T, error),
	attrs ...attribute.KeyValue,
) (T, error) {
	child, done := RawSpan(parent, name, attrs...)
	result, err := fn(child.(C))
	done(err)

	return result, err
}

// SpanErr wraps a typed-context error function with automatic span lifecycle.
// This is the error-only variant of [Span] — use when the function returns only
// an error (no result value).
//
//	func (s *service) Advance(ctx gamectx.GameContext) error {
//	    return observe.SpanErr(ctx, "game.advance",
//	        func(ctx gamectx.GameContext) error {
//	            // ctx carries the child span — compile-time safe
//	        },
//	    )
//	}
//
//nolint:forcetypeassert // Rebaseable.Rebase contract guarantees type preservation
func SpanErr[C ctx.Rebaseable](
	parent C,
	name string,
	fn func(C) error,
	attrs ...attribute.KeyValue,
) error {
	child, done := RawSpan(parent, name, attrs...)
	err := fn(child.(C))
	done(err)

	return err
}
