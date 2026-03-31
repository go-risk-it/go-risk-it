package service

import (
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
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
	querier db.Querier,
	move T,
) (R, error) {
	return observe.SpanFunc(
		gameCtx,
		"game.move.perform",
		func(gameCtx ctx.GameContext) (R, error) {
			return t.inner.Perform(gameCtx, querier, move)
		},
		attribute.String("phase", string(t.inner.PhaseType())),
	)
}

func (t *tracedService[T, R]) Walk(
	gameCtx ctx.GameContext,
	querier db.Querier,
	voluntaryAdvancement bool,
) (sqlc.GamePhaseType, error) {
	return observe.SpanFunc(
		gameCtx,
		"game.move.walk",
		func(gameCtx ctx.GameContext) (sqlc.GamePhaseType, error) {
			return t.inner.Walk(gameCtx, querier, voluntaryAdvancement)
		},
		attribute.String("phase", string(t.inner.PhaseType())),
	)
}

func (t *tracedService[T, R]) Advance(
	gameCtx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	performResult R,
) error {
	return observe.SpanErr(gameCtx, "game.move.advance", func(gameCtx ctx.GameContext) error {
		return t.inner.Advance(gameCtx, querier, targetPhase, performResult)
	}, attribute.String("phase", string(t.inner.PhaseType())))
}

func (t *tracedService[T, R]) PhaseType() sqlc.GamePhaseType {
	return t.inner.PhaseType()
}
