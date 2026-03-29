package service

import (
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	gameCtx, span := tracing.StartGameSpan(gameCtx, "game.move.perform",
		attribute.String("phase", string(t.inner.PhaseType())),
	)
	defer span.End()

	result, err := t.inner.Perform(gameCtx, querier, move)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return result, err
}

func (t *tracedService[T, R]) Walk(
	gameCtx ctx.GameContext,
	querier db.Querier,
	voluntaryAdvancement bool,
) (sqlc.GamePhaseType, error) {
	gameCtx, span := tracing.StartGameSpan(gameCtx, "game.move.walk",
		attribute.String("phase", string(t.inner.PhaseType())),
	)
	defer span.End()

	targetPhase, err := t.inner.Walk(gameCtx, querier, voluntaryAdvancement)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return targetPhase, err
}

func (t *tracedService[T, R]) Advance(
	gameCtx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	performResult R,
) error {
	gameCtx, span := tracing.StartGameSpan(gameCtx, "game.move.advance",
		attribute.String("phase", string(t.inner.PhaseType())),
	)
	defer span.End()

	err := t.inner.Advance(gameCtx, querier, targetPhase, performResult)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *tracedService[T, R]) PhaseType() sqlc.GamePhaseType {
	return t.inner.PhaseType()
}
