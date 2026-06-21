package orchestration

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

// PhaseAdvancer is the ISP-clean interface for voluntary phase advancement.
// It is separate from Orchestrator[T,R] because advancement controllers
// do not need the full move orchestration contract.
type PhaseAdvancer interface {
	AdvancePhase(ctx gamectx.GameContext) error
}

// AttackPhaseAdvancer is a wrapper type for FX disambiguation.
// It allows fx.Provide to distinguish AttackPhaseAdvancer from ReinforcePhaseAdvancer.
type AttackPhaseAdvancer struct{ PhaseAdvancer }

// ReinforcePhaseAdvancer is a wrapper type for FX disambiguation.
type ReinforcePhaseAdvancer struct{ PhaseAdvancer }

// CardsPhaseAdvancer is a wrapper type for FX disambiguation.
type CardsPhaseAdvancer struct{ PhaseAdvancer }

// AdvancePhase performs a voluntary phase advancement using the same
// effects-first pipeline as OrchestrateMove, minus Perform and LogMove.
// It emits MoveCompleted so downstream handlers (state broadcaster,
// headlines detector) observe the phase change.
func (s *orchestrator[T, R]) AdvancePhase(ctx gamectx.GameContext) error {
	return observe.SpanErr(ctx, "game.advance_phase", func(ctx gamectx.GameContext) error {
		unlock := s.gameLocks.Lock(ctx.GameID())
		defer unlock()

		outcome, err := s.advancePhasePipeline(ctx)
		if err != nil {
			return fmt.Errorf("unable to advance phase: %w", err)
		}

		s.stateStore.Store(ctx.GameID(), outcome.newState)

		s.emitMoveCompleted(ctx, outcome)

		return nil
	}, attribute.String("phase", string(s.service.PhaseType())))
}

// advancePhasePipeline runs the effects-first pipeline for voluntary advancement:
// cache-get → validate → Walk(voluntary=true) → Advance(zero R) →
// BuildNewState → buildPersistenceEffect → Persist(write-only TX).
func (s *orchestrator[T, R]) advancePhasePipeline(
	ctx gamectx.GameContext,
) (moveOutcome[R], error) {
	phase := s.service.PhaseType()

	// 1. Get cached state (or warm from DB).
	prevState, err := s.getOrWarmState(ctx, phase)
	if err != nil {
		return moveOutcome[R]{}, err
	}

	gameState := GameStateFromCache(prevState)

	// 2. Validate.
	if err := s.validationService.Validate(
		ctx, gameState, prevState.PublicSnapshot.Players,
	); err != nil {
		return moveOutcome[R]{}, fmt.Errorf("invalid advance: %w", err)
	}

	// 3. Walk with voluntary=true and an empty MoveEffect.
	emptyEffect := moveservice.MoveEffect{}
	walkCtx := buildWalkContext(prevState, emptyEffect, true, ctx.UserID())

	targetPhase, err := s.service.Walk(walkCtx)
	if err != nil {
		return moveOutcome[R]{}, fmt.Errorf("unable to walk phase: %w", err)
	}

	// 4. Advance.
	advEffect, err := s.advanceToTarget(ctx, targetPhase, prevState, emptyEffect)
	if err != nil {
		return moveOutcome[R]{}, err
	}

	// 5-7. Build state, build persistence, persist.
	return s.buildAndPersistAdvance(
		ctx, prevState, gameState.Turn, phase,
		emptyEffect, advEffect, targetPhase,
	)
}

// buildAndPersistAdvance handles the tail of the advance pipeline:
// BuildNewState → buildPersistenceEffect → Persist.
func (s *orchestrator[T, R]) buildAndPersistAdvance(
	ctx gamectx.GameContext,
	prevState *snapshot.CachedGameState,
	turn int64,
	phase sqlc.GamePhaseType,
	emptyEffect moveservice.MoveEffect,
	advEffect moveservice.AdvanceEffect,
	targetPhase sqlc.GamePhaseType,
) (moveOutcome[R], error) {
	newState := BuildNewState(
		prevState,
		&emptyEffect,
		&advEffect,
		targetPhase,
		"", // no winner — advancements never end the game
	)

	moveCtx := MoveContext{
		gameID:    ctx.GameID(),
		userID:    ctx.UserID(),
		phaseType: phase,
	}

	persistEffect := buildPersistenceEffect(
		moveCtx, nil, &advEffect, prevState, targetPhase, false,
	)

	if err := Persist(ctx, s.querier, s.phaseService, persistEffect); err != nil {
		return moveOutcome[R]{}, fmt.Errorf("unable to persist advance: %w", err)
	}

	observe.SpanEvent(ctx, "game.advance.persisted")

	var zero R

	return moveOutcome[R]{
		targetPhase: targetPhase,
		gameOver:    false,
		result:      zero,
		turn:        turn,
		newState:    newState,
		prevRegions: prevState.PublicSnapshot.Regions,
	}, nil
}

func (s *orchestrator[T, R]) advanceToTarget(
	ctx gamectx.GameContext,
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

	advEffect, err := s.service.Advance(ctx, targetPhase, zero, advCtx)
	if err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("unable to advance phase: %w", err)
	}

	return advEffect, nil
}

// NewAttackPhaseAdvancer creates an AttackPhaseAdvancer from the existing
// AttackOrchestrator. No type assertion needed — Orchestrator[T,R] embeds PhaseAdvancer.
func NewAttackPhaseAdvancer(orch AttackOrchestrator) AttackPhaseAdvancer {
	return AttackPhaseAdvancer{PhaseAdvancer: orch}
}

// NewReinforcePhaseAdvancer creates a ReinforcePhaseAdvancer from the existing
// ReinforceOrchestrator. No type assertion needed — Orchestrator[T,R] embeds PhaseAdvancer.
func NewReinforcePhaseAdvancer(orch ReinforceOrchestrator) ReinforcePhaseAdvancer {
	return ReinforcePhaseAdvancer{PhaseAdvancer: orch}
}

// NewCardsPhaseAdvancer creates a CardsPhaseAdvancer from the existing
// CardsOrchestrator. No type assertion needed — Orchestrator[T,R] embeds PhaseAdvancer.
func NewCardsPhaseAdvancer(orch CardsOrchestrator) CardsPhaseAdvancer {
	return CardsPhaseAdvancer{PhaseAdvancer: orch}
}
