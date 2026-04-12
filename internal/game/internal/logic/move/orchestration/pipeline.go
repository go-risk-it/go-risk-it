package orchestration

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

func (s *orchestrator[T, R]) OrchestrateMove(
	ctx gamectx.GameContext,
	move T,
) error {
	return observe.SpanErr(ctx, "game.orchestrate_move", func(ctx gamectx.GameContext) error {
		unlock := s.gameLocks.Lock(ctx.GameID())
		defer unlock()

		outcome, err := s.orchestrateMovePipeline(ctx, move)
		if err != nil {
			return fmt.Errorf("unable to perform move: %w", err)
		}

		s.stateStore.Store(ctx.GameID(), outcome.newState)

		s.emitMoveCompleted(ctx, outcome)

		return nil
	}, attribute.String("phase", string(s.service.PhaseType())))
}

// orchestrateMovePipeline runs the effects-first pipeline:
// cache-get → validate → perform(pure) → walk → advance(pure) → checkMission(pure)
// → BuildNewState → buildPersistenceEffect → Persist(write-only TX).
func (s *orchestrator[T, R]) orchestrateMovePipeline(
	ctx gamectx.GameContext,
	move T,
) (moveOutcome[R], error) {
	phase := s.service.PhaseType()

	// 1. Get cached state (or warm from DB — direct querier read, no TX).
	prevState, err := s.getOrWarmState(ctx, phase)
	if err != nil {
		return moveOutcome[R]{}, err
	}

	gameState := GameStateFromCache(prevState)

	// 2. Validate move (pure, from cache).
	if err := s.validationService.Validate(
		ctx, gameState, prevState.PublicSnapshot.Players,
	); err != nil {
		return moveOutcome[R]{}, fmt.Errorf("invalid move: %w", err)
	}

	observe.SpanEvent(ctx, "game.move.validated")

	// 3. Perform (pure, no querier).
	performResult, effect, err := s.service.Perform(ctx, move, prevState)
	if err != nil {
		return moveOutcome[R]{}, fmt.Errorf("unable to perform move: %w", err)
	}

	observe.SpanEvent(ctx, "game.move.performed")

	// 4-5. Resolve target phase: checkMission → walk → advance (all pure).
	targetPhase, gameOver, advEffect, err := s.resolveTargetPhase(
		ctx, performResult, prevState, effect,
	)
	if err != nil {
		return moveOutcome[R]{}, err
	}

	// 6-8. Build state, build persistence, persist, emit.
	return s.buildAndPersistMove(
		ctx, prevState, gameState.Turn, phase,
		performResult, effect, advEffect, targetPhase, gameOver, move,
	)
}

// buildAndPersistMove handles the tail of the move pipeline:
// BuildNewState → buildPersistenceEffect → Persist → game-over metrics.
func (s *orchestrator[T, R]) buildAndPersistMove(
	ctx gamectx.GameContext,
	prevState *snapshot.CachedGameState,
	turn int64,
	phase sqlc.GamePhaseType,
	performResult R,
	effect moveservice.MoveEffect,
	advEffect moveservice.AdvanceEffect,
	targetPhase sqlc.GamePhaseType,
	gameOver bool,
	move T,
) (moveOutcome[R], error) {
	winnerUserID := ""
	if gameOver {
		winnerUserID = ctx.UserID()
	}

	var advEffectPtr *moveservice.AdvanceEffect
	if targetPhase != s.service.PhaseType() && !gameOver {
		advEffectPtr = &advEffect
	}

	newState := BuildNewState(
		prevState,
		&effect,
		advEffectPtr,
		targetPhase,
		winnerUserID,
	)

	moveCtx := MoveContext{
		gameID:    ctx.GameID(),
		userID:    ctx.UserID(),
		phaseType: phase,
		moveData:  move,
		result:    performResult,
	}

	persistEffect := buildPersistenceEffect(
		moveCtx, &effect, advEffectPtr, prevState, gameOver,
	)

	if err := Persist(ctx, s.querier, s.phaseService, persistEffect); err != nil {
		return moveOutcome[R]{}, fmt.Errorf("unable to persist move: %w", err)
	}

	observe.SpanEvent(ctx, "game.move.persisted")

	if gameOver {
		observe.SpanEvent(ctx, "game_won")
		s.recordGameFinished(ctx)
	}

	return moveOutcome[R]{
		targetPhase: targetPhase,
		gameOver:    gameOver,
		result:      performResult,
		turn:        turn,
		newState:    newState,
		prevRegions: prevState.PublicSnapshot.Regions,
	}, nil
}

// getOrWarmState returns the cached state or warms from DB on cache miss.
// Under the per-game mutex, a direct querier read is safe (no TX needed).
func (s *orchestrator[T, R]) getOrWarmState(
	ctx gamectx.GameContext,
	expectedPhase sqlc.GamePhaseType,
) (*snapshot.CachedGameState, error) {
	// Fast path: cache hit.
	if cached := s.stateStore.Get(ctx.GameID()); cached != nil {
		gameState := GameStateFromCache(cached)

		if gameState.Phase != expectedPhase {
			return nil, domainerrors.NewConflictErrorf(
				"game is in phase %s, expected %s",
				gameState.Phase, expectedPhase,
			)
		}

		return cached, nil
	}

	// Slow path: cache miss — warm from DB using direct querier (no TX).
	return s.warmFromDB(ctx, expectedPhase)
}

// warmFromDB reads the full game state from the database using a direct querier
// (no transaction). This is safe because we hold the per-game mutex.
func (s *orchestrator[T, R]) warmFromDB(
	ctx gamectx.GameContext,
	expectedPhase sqlc.GamePhaseType,
) (*snapshot.CachedGameState, error) {
	gameState, err := s.gameService.GetGameStateWithQuerier(ctx, s.querier)
	if err != nil {
		return nil, fmt.Errorf("unable to get game state: %w", err)
	}

	if gameState.Phase != expectedPhase {
		return nil, domainerrors.NewConflictErrorf(
			"game is in phase %s, expected %s",
			gameState.Phase, expectedPhase,
		)
	}

	prevState, err := s.getOrWarmPrevState(ctx, s.querier, gameState.Turn)
	if err != nil {
		return nil, fmt.Errorf("unable to warm state: %w", err)
	}

	return prevState, nil
}

func (s *orchestrator[T, R]) resolveTargetPhase(
	ctx gamectx.GameContext,
	performResult R,
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
) (sqlc.GamePhaseType, bool, moveservice.AdvanceEffect, error) {
	if accomplished, err := s.checkMission(
		ctx, prevState, effect,
	); err != nil {
		return "", false, moveservice.AdvanceEffect{}, err
	} else if accomplished {
		return s.service.PhaseType(), true, moveservice.AdvanceEffect{}, nil
	}

	// Universal domination: if the player owns all regions after the move,
	// the game is won regardless of their specific mission.
	if IsDomination(prevState, effect, ctx.UserID()) {
		return s.service.PhaseType(), true, moveservice.AdvanceEffect{}, nil
	}

	targetPhase, advEffect, err := s.walkAndAdvance(
		ctx, performResult, prevState, effect,
	)
	if err != nil {
		return "", false, moveservice.AdvanceEffect{}, err
	}

	return targetPhase, false, advEffect, nil
}

func (s *orchestrator[T, R]) walkAndAdvance(
	ctx gamectx.GameContext,
	performResult R,
	prevState *snapshot.CachedGameState,
	effect moveservice.MoveEffect,
) (sqlc.GamePhaseType, moveservice.AdvanceEffect, error) {
	walkCtx := buildWalkContext(prevState, effect, false, ctx.UserID())

	targetPhase, err := s.service.Walk(walkCtx)
	if err != nil {
		return "", moveservice.AdvanceEffect{}, fmt.Errorf("unable to walk phase: %w", err)
	}

	if targetPhase == s.service.PhaseType() {
		return targetPhase, moveservice.AdvanceEffect{}, nil
	}

	advCtx, err := s.buildAdvanceContext(ctx, prevState, effect, ctx.UserID())
	if err != nil {
		return "", moveservice.AdvanceEffect{}, fmt.Errorf(
			"unable to build advance context: %w",
			err,
		)
	}

	advEffect, err := s.service.Advance(
		ctx, targetPhase, performResult, advCtx,
	)
	if err != nil {
		return "", moveservice.AdvanceEffect{}, fmt.Errorf("unable to advance move: %w", err)
	}

	return targetPhase, advEffect, nil
}
