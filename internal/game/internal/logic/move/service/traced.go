package service

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
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

// performResult bundles the two non-error return values of Perform so that
// observe.Span (which accepts a single result type) can wrap the call.
type performResult[R any] struct {
	result R
	effect MoveEffect
}

func (t *tracedService[T, R]) Perform(
	gameCtx ctx.GameContext,
	querier db.Querier,
	move T,
	prev *snapshot.CachedGameState,
) (R, MoveEffect, error) {
	pr, err := observe.Span(
		gameCtx,
		"game.move.perform",
		func(gameCtx ctx.GameContext) (performResult[R], error) {
			r, eff, err := t.inner.Perform(gameCtx, querier, move, prev)

			return performResult[R]{result: r, effect: eff}, err
		},
		attribute.String("phase", string(t.inner.PhaseType())),
	)

	return pr.result, pr.effect, err
}

// Walk is a pure function (no context, no DB) — tracing is not applicable.
func (t *tracedService[T, R]) Walk(wctx WalkContext) (sqlc.GamePhaseType, error) {
	return t.inner.Walk(wctx)
}

func (t *tracedService[T, R]) Advance(
	gameCtx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	performResult R,
	advCtx AdvanceContext,
) (AdvanceEffect, error) {
	return observe.Span(
		gameCtx,
		"game.move.advance",
		func(gameCtx ctx.GameContext) (AdvanceEffect, error) {
			return t.inner.Advance(gameCtx, querier, targetPhase, performResult, advCtx)
		},
		attribute.String("phase", string(t.inner.PhaseType())),
	)
}

func (t *tracedService[T, R]) PhaseType() sqlc.GamePhaseType {
	return t.inner.PhaseType()
}
