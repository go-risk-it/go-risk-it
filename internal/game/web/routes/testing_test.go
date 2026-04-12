package routes_test

import (
	"context"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"go.opentelemetry.io/otel/trace/noop"
)

func testGameContext() gamectx.GameContext {
	uc := ctx.WithUserID(
		ctx.WithSpan(context.Background(), noop.Span{}),
		"user-123",
	)

	return gamectx.WithGameID(uc, 42)
}
