package advancement

import (
	"fmt"
	"log/slog"
	"time"

	gamectx "github.com/go-risk-it/go-risk-it/internal/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/events"
	gameevt "github.com/go-risk-it/go-risk-it/internal/events/game"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/validation"
	moveservice "github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/go-risk-it/go-risk-it/internal/tracing"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type Service[T, R any] interface {
	Advance(ctx gamectx.GameContext) error
	AdvanceWithQuerier(
		ctx gamectx.GameContext,
		querier db.Querier,
	) (sqlc.GamePhaseType, error)
}

// advanceOutcome carries the result of advanceInternal for post-commit side effects.
type advanceOutcome struct {
	targetPhase sqlc.GamePhaseType
	turn        int64
}

type service[T, R any] struct {
	querier           db.Querier
	gameState         state.Service
	moveService       moveservice.Service[T, R]
	validationService validation.Service
	bus               events.Bus
	metrics           *metrics.Metrics
}

func NewService[T, R any](
	gameState state.Service,
	querier db.Querier,
	moveService moveservice.Service[T, R],
	validationService validation.Service,
	bus events.Bus,
	metrics *metrics.Metrics,
) Service[T, R] {
	return &service[T, R]{
		gameState:         gameState,
		querier:           querier,
		moveService:       moveService,
		validationService: validationService,
		bus:               bus,
		metrics:           metrics,
	}
}

func (s *service[T, R]) Advance(ctx gamectx.GameContext) error {
	currentPhase := s.moveService.PhaseType()

	ctx, span := tracing.StartGameSpan(ctx, "game.advance",
		attribute.String("phase", string(currentPhase)),
	)
	defer span.End()

	outcome, err := dbutil.InTransactionWithIsolation(
		s.querier,
		ctx,
		s.metrics,
		pgx.RepeatableRead,
		func(q db.Querier) (advanceOutcome, error) {
			return s.advanceInternal(ctx, q)
		},
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return fmt.Errorf("unable to perform move: %w", err)
	}

	s.bus.Emit(ctx, gameevt.NewPhaseTransitioned(
		ctx.GameID(),
		ctx.UserID(),
		time.Now(),
		currentPhase,
		outcome.targetPhase,
		outcome.turn,
	))

	return nil
}

func (s *service[T, R]) AdvanceWithQuerier(
	ctx gamectx.GameContext,
	querier db.Querier,
) (sqlc.GamePhaseType, error) {
	outcome, err := s.advanceInternal(ctx, querier)
	if err != nil {
		return "", err
	}

	return outcome.targetPhase, nil
}

func (s *service[T, R]) advanceInternal(
	ctx gamectx.GameContext,
	querier db.Querier,
) (advanceOutcome, error) {
	currentPhase := s.moveService.PhaseType()
	phase := string(currentPhase)

	slog.InfoContext(ctx, "processing request to advance phase",
		"currentPhase", currentPhase,
	)

	game, err := s.getAndValidateState(ctx, querier, phase)
	if err != nil {
		return advanceOutcome{}, err
	}

	if game.Phase != currentPhase {
		return advanceOutcome{}, domainerrors.NewConflictErrorf(
			"game is in phase %s, expected %s",
			game.Phase,
			currentPhase,
		)
	}

	targetPhase, err := s.walkAndAdvance(ctx, querier, phase, currentPhase)
	if err != nil {
		return advanceOutcome{}, err
	}

	return advanceOutcome{
		targetPhase: targetPhase,
		turn:        game.Turn,
	}, nil
}

func (s *service[T, R]) getAndValidateState(
	ctx gamectx.GameContext,
	querier db.Querier,
	phase string,
) (*state.Game, error) {
	var game *state.Game

	if err := tracing.SpanStep(
		ctx, "game.advance.get_state", phase,
		func(spanCtx gamectx.GameContext) error {
			var getErr error
			game, getErr = s.gameState.GetGameStateWithQuerier(
				spanCtx, querier,
			)

			return getErr
		},
	); err != nil {
		return nil, fmt.Errorf("unable to get game state: %w", err)
	}

	if err := tracing.SpanStep(
		ctx, "game.advance.validate", phase,
		func(spanCtx gamectx.GameContext) error {
			return s.validationService.Validate(
				spanCtx, querier, game,
			)
		},
	); err != nil {
		slog.ErrorContext(ctx, "validation failed", "error", err)

		return nil, fmt.Errorf("validation failed: %w", err)
	}

	slog.DebugContext(ctx, "game is in phase", "phase", game.Phase)

	return game, nil
}

func (s *service[T, R]) walkAndAdvance(
	ctx gamectx.GameContext,
	querier db.Querier,
	phase string,
	currentPhase sqlc.GamePhaseType,
) (sqlc.GamePhaseType, error) {
	var targetPhase sqlc.GamePhaseType

	if err := tracing.SpanStep(
		ctx, "game.advance.walk", phase,
		func(spanCtx gamectx.GameContext) error {
			var walkErr error
			targetPhase, walkErr = s.moveService.Walk(
				spanCtx, querier, true,
			)

			return walkErr
		},
	); err != nil {
		return "", fmt.Errorf("unable to walk to target phase: %w", err)
	}

	if err := tracing.SpanStep(
		ctx, "game.advance.advance", phase,
		func(spanCtx gamectx.GameContext) error {
			var zero R

			return s.moveService.Advance(
				spanCtx, querier, targetPhase, zero,
			)
		},
	); err != nil {
		return "", fmt.Errorf("unable to perform move: %w", err)
	}

	slog.InfoContext(ctx, "phase advanced successfully",
		"from", currentPhase,
	)

	return targetPhase, nil
}
