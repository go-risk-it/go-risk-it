package service

import (
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/tracing"
	"go.opentelemetry.io/otel/attribute"
)

type tracedService[T, R any] struct {
	inner Service[T, R]
}

func NewTracedService[T, R any](inner Service[T, R]) Service[T, R] {
	return &tracedService[T, R]{inner: inner}
}

//nolint:nonamedreturns // named returns needed for defer-based error recording
func (t *tracedService[T, R]) Perform(
	gameCtx ctx.GameContext,
	querier db.Querier,
	move T,
) (result R, err error) {
	gameCtx, done := tracing.StartGameSpan(gameCtx, "game.move.perform",
		attribute.String("phase", string(t.inner.PhaseType())),
	)
	defer func() { done(err) }()

	return t.inner.Perform(gameCtx, querier, move)
}

//nolint:nonamedreturns // named returns needed for defer-based error recording
func (t *tracedService[T, R]) Walk(
	gameCtx ctx.GameContext,
	querier db.Querier,
	voluntaryAdvancement bool,
) (targetPhase sqlc.GamePhaseType, err error) {
	gameCtx, done := tracing.StartGameSpan(gameCtx, "game.move.walk",
		attribute.String("phase", string(t.inner.PhaseType())),
	)
	defer func() { done(err) }()

	return t.inner.Walk(gameCtx, querier, voluntaryAdvancement)
}

//nolint:nonamedreturns // named returns needed for defer-based error recording
func (t *tracedService[T, R]) Advance(
	gameCtx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	performResult R,
) (err error) {
	gameCtx, done := tracing.StartGameSpan(gameCtx, "game.move.advance",
		attribute.String("phase", string(t.inner.PhaseType())),
	)
	defer func() { done(err) }()

	return t.inner.Advance(gameCtx, querier, targetPhase, performResult)
}

func (t *tracedService[T, R]) PhaseType() sqlc.GamePhaseType {
	return t.inner.PhaseType()
}
