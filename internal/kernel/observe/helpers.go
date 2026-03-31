package observe

import (
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"go.opentelemetry.io/otel/attribute"
)

// TypedSpan starts a new child span and returns the enriched context with its
// original type preserved, plus a done function. The type parameter C is inferred
// from the parent's static type — callers never need to specify it explicitly.
//
// This is the preferred span creation function for typed contexts (GameContext,
// LobbyContext). For plain context.Context, use [Span] instead.
//
//	ctx, done := observe.TypedSpan(gameCtx, "game.move.validate")
//	defer func() { done(err) }()
//	// ctx is still GameContext — compile-time safe
//

func TypedSpan[C ctx.Rebaseable](
	parent C,
	name string,
	attrs ...attribute.KeyValue,
) (C, func(error)) {
	child, done := Span(parent, name, attrs...)
	//nolint:forcetypeassert // Rebaseable.Rebase contract guarantees type preservation
	return child.(C), done
}

// TypedSpanFunc wraps a typed-context function with automatic span lifecycle.
// The context type C is preserved through auto-rebase — fn receives the same
// typed context as parent. Eliminates both done(nil) bugs and type assertions.
//
//nolint:forcetypeassert // Rebaseable.Rebase contract guarantees type preservation
func TypedSpanFunc[C ctx.Rebaseable, T any](
	parent C,
	name string,
	fn func(C) (T, error),
	attrs ...attribute.KeyValue,
) (T, error) {
	child, done := Span(parent, name, attrs...)
	result, err := fn(child.(C))
	done(err)

	return result, err
}

// TypedSpanErr wraps a typed-context error function with automatic span lifecycle.
//
//nolint:forcetypeassert // Rebaseable.Rebase contract guarantees type preservation
func TypedSpanErr[C ctx.Rebaseable](
	parent C,
	name string,
	fn func(C) error,
	attrs ...attribute.KeyValue,
) error {
	child, done := Span(parent, name, attrs...)
	err := fn(child.(C))
	done(err)

	return err
}
