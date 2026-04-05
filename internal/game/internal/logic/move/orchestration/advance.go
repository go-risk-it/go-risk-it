package orchestration

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/state"
	dbutil "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
)

// PhaseAdvancer is the ISP-clean interface for voluntary phase advancement.
// It is separate from Orchestrator[T,R] because advancement controllers
// do not need the full move orchestration contract.
type PhaseAdvancer interface {
	AdvancePhase(ctx gamectx.GameContext) error
}

// Wrapper types for FX disambiguation — each is a distinct Go type so that
// fx.Provide can distinguish AttackPhaseAdvancer from ReinforcePhaseAdvancer.
type AttackPhaseAdvancer struct{ PhaseAdvancer }

type ReinforcePhaseAdvancer struct{ PhaseAdvancer }

type CardsPhaseAdvancer struct{ PhaseAdvancer }

// AdvancePhase performs a voluntary phase advancement using the same
// transactional pipeline as OrchestrateMove, minus Perform and LogMove.
// It emits MoveCompleted so downstream handlers (state broadcaster,
// headlines detector) observe the phase change.
func (s *orchestrator[T, R]) AdvancePhase(ctx gamectx.GameContext) error {
	return observe.SpanErr(ctx, "game.advance_phase", func(ctx gamectx.GameContext) error {
		outcome, err := s.executeAdvanceTransaction(ctx)
		if err != nil {
			return fmt.Errorf("unable to advance phase: %w", err)
		}

		s.stateStore.Store(ctx.GameID(), outcome.newState)

		s.emitMoveCompleted(ctx, outcome)

		return nil
	}, attribute.String("phase", string(s.service.PhaseType())))
}

// executeAdvanceTransaction runs the advancement in a RepeatableRead
// transaction: validate state, Walk(voluntary=true), Advance(zero R),
// BuildNewState.
func (s *orchestrator[T, R]) executeAdvanceTransaction(
	ctx gamectx.GameContext,
) (moveOutcome[R], error) {
	return dbutil.InTransactionWithIsolation(
		s.querier, ctx, s.stateMetrics, pgx.RepeatableRead,
		func(querier db.Querier) (moveOutcome[R], error) {
			phase := s.service.PhaseType()

			gameState, err := s.gameService.GetGameStateWithQuerier(ctx, querier)
			if err != nil {
				return moveOutcome[R]{}, fmt.Errorf(
					"unable to get game state: %w", err,
				)
			}

			if gameState.Phase != phase {
				return moveOutcome[R]{}, domainerrors.NewConflictErrorf(
					"game is in phase %s, expected %s",
					gameState.Phase, phase,
				)
			}

			if err := s.validationService.Validate(ctx, querier, gameState); err != nil {
				return moveOutcome[R]{}, fmt.Errorf("invalid advance: %w", err)
			}

			return s.advancePhaseWithQuerier(ctx, querier, gameState)
		},
	)
}

func (s *orchestrator[T, R]) advancePhaseWithQuerier(
	ctx gamectx.GameContext,
	querier db.Querier,
	gameState *state.Game,
) (moveOutcome[R], error) {
	turn := gameState.Turn

	prevState, err := s.getOrWarmPrevState(ctx, querier, turn)
	if err != nil {
		return moveOutcome[R]{}, fmt.Errorf("unable to get previous state: %w", err)
	}

	// Voluntary advancement: Walk with voluntary=true and an empty MoveEffect.
	emptyEffect := moveservice.MoveEffect{}
	walkCtx := buildWalkContext(prevState, emptyEffect, true, ctx.UserID())

	targetPhase, err := s.service.Walk(walkCtx)
	if err != nil {
		return moveOutcome[R]{}, fmt.Errorf("unable to walk phase: %w", err)
	}

	advEffect, err := s.advanceToTarget(ctx, querier, targetPhase, prevState, emptyEffect)
	if err != nil {
		return moveOutcome[R]{}, err
	}

	return s.buildAdvanceOutcome(
		ctx, prevState, emptyEffect, advEffect, targetPhase, turn,
	), nil
}

func (s *orchestrator[T, R]) advanceToTarget(
	ctx gamectx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
) (moveservice.AdvanceEffect, error) {
	advCtx, err := s.buildAdvanceContext(ctx, prevState, effect, ctx.UserID())
	if err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf(
			"unable to build advance context: %w", err,
		)
	}

	var zero R

	advEffect, err := s.service.Advance(ctx, querier, targetPhase, zero, advCtx)
	if err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("unable to advance phase: %w", err)
	}

	return advEffect, nil
}

// buildAdvanceOutcome assembles the moveOutcome for an advancement. Unlike
// buildOutcome, there is no Perform result and no move log.
func (s *orchestrator[T, R]) buildAdvanceOutcome(
	ctx gamectx.GameContext,
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
	advEffect moveservice.AdvanceEffect,
	targetPhase sqlc.GamePhaseType,
	turn int64,
) moveOutcome[R] {
	newState := BuildNewState(
		prevState,
		&effect,
		&advEffect,
		targetPhase,
		false, // advancements never end the game
		"",    // no winner
	)

	var zero R

	return moveOutcome[R]{
		targetPhase: targetPhase,
		gameOver:    false,
		result:      zero,
		moveLog:     sqlc.GameMoveLog{}, // no move log for advancements
		turn:        turn,
		newState:    newState,
		prevRegions: prevState.PublicSnapshot.Regions,
	}
}

// NewAttackPhaseAdvancer creates an AttackPhaseAdvancer from the existing
// AttackOrchestrator.
func NewAttackPhaseAdvancer(orch AttackOrchestrator) AttackPhaseAdvancer {
	return AttackPhaseAdvancer{PhaseAdvancer: orch.(PhaseAdvancer)}
}

// NewReinforcePhaseAdvancer creates a ReinforcePhaseAdvancer from the existing
// ReinforceOrchestrator.
func NewReinforcePhaseAdvancer(orch ReinforceOrchestrator) ReinforcePhaseAdvancer {
	return ReinforcePhaseAdvancer{PhaseAdvancer: orch.(PhaseAdvancer)}
}

// NewCardsPhaseAdvancer creates a CardsPhaseAdvancer from the existing
// CardsOrchestrator.
func NewCardsPhaseAdvancer(orch CardsOrchestrator) CardsPhaseAdvancer {
	return CardsPhaseAdvancer{PhaseAdvancer: orch.(PhaseAdvancer)}
}
