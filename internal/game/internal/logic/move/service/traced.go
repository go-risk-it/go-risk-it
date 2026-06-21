package service

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

type tracedService[T, R any] struct {
	inner Service[T, R]
}

func NewTracedService[T, R any](inner Service[T, R]) Service[T, R] {
	return &tracedService[T, R]{inner: inner}
}

func (t *tracedService[T, R]) Perform(
	gameCtx ctx.GameContext,
	move T,
	prev *snapshot.CachedGameState,
) (R, MoveEffect, error) {
	observe.SpanEvent(gameCtx, "game.move.perform",
		attribute.String("phase", string(t.inner.PhaseType())))

	return t.inner.Perform(gameCtx, move, prev)
}

// Walk is a pure function (no context, no DB) — tracing is not applicable.
func (t *tracedService[T, R]) Walk(wctx WalkContext) (sqlc.GamePhaseType, error) {
	return t.inner.Walk(wctx)
}

func (t *tracedService[T, R]) Advance(
	gameCtx ctx.GameContext,
	targetPhase sqlc.GamePhaseType,
	performResult R,
	advCtx AdvanceContext,
) (AdvanceEffect, error) {
	observe.SpanEvent(gameCtx, "game.move.advance",
		attribute.String("phase", string(t.inner.PhaseType())))

	return t.inner.Advance(gameCtx, targetPhase, performResult, advCtx)
}

func (t *tracedService[T, R]) PhaseType() sqlc.GamePhaseType {
	return t.inner.PhaseType()
}
