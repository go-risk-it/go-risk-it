package advancement

import (
	"fmt"
	"time"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/orchestration"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/game/tracing"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	dbutil "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
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
	validationService orchestration.ValidationService
	bus               eventbus.Publisher
	metrics           *metrics.StateMetrics
}

func NewService[T, R any](
	gameState state.Service,
	querier db.Querier,
	moveService moveservice.Service[T, R],
	validationService orchestration.ValidationService,
	bus eventbus.Publisher,
	metrics *metrics.StateMetrics,
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

//nolint:nonamedreturns // named returns needed for defer-based error recording
func (s *service[T, R]) Advance(ctx gamectx.GameContext) (err error) {
	currentPhase := s.moveService.PhaseType()

	ctx, done := tracing.StartGameSpan(ctx, "game.advance",
		attribute.String("phase", string(currentPhase)),
	)
	defer func() { done(err) }()

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

	game, err := s.getAndValidateState(ctx, querier)
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

	targetPhase, err := s.walkAndAdvance(ctx, querier)
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
) (*state.Game, error) {
	game, err := s.gameState.GetGameStateWithQuerier(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("unable to get game state: %w", err)
	}

	if err := s.validationService.Validate(ctx, querier, game); err != nil {
		observe.Error(ctx, err, "validation failed")

		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return game, nil
}

func (s *service[T, R]) walkAndAdvance(
	ctx gamectx.GameContext,
	querier db.Querier,
) (sqlc.GamePhaseType, error) {
	targetPhase, err := s.moveService.Walk(ctx, querier, true)
	if err != nil {
		return "", fmt.Errorf("unable to walk to target phase: %w", err)
	}

	var zero R

	if err := s.moveService.Advance(
		ctx, querier, targetPhase, zero,
	); err != nil {
		return "", fmt.Errorf("unable to perform move: %w", err)
	}

	return targetPhase, nil
}
