package tracing

import (
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

// StartGameSpan starts a new span and threads it through the GameContext chain.
// The caller must invoke the returned done function when the operation completes:
//
//	gameCtx, done := tracing.StartGameSpan(gameCtx, "name", attrs...)
//	defer done(nil)          // no error to report
//	defer func() { done(err) }()  // report named-return error
func StartGameSpan(
	gameCtx ctx.GameContext,
	spanName string,
	attrs ...attribute.KeyValue,
) (ctx.GameContext, func(error)) {
	plainCtx, done := observe.Span(gameCtx, spanName, attrs...)

	return gameCtx.WithBase(plainCtx), done
}
