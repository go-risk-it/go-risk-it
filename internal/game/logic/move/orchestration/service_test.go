package orchestration_test

import (
	"context"
	"errors"
	"testing"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	gamemetrics "github.com/go-risk-it/go-risk-it/internal/game/logic/metrics"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/reinforce"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	mockdb "github.com/go-risk-it/go-risk-it/mocks/internal_/game/data/db"
	mockmission "github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/mission"
	mockorchestration "github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/move/orchestration"
	mockservice "github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/move/service"
	mockstate "github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/state"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

const (
	testGameID = int64(42)
	testUserID = "player1"
	testTurn   = int64(5)
)

func testCtx() gamectx.GameContext {
	traceCtx := kernelctx.WithSpan(context.Background(), tracenoop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, testUserID)

	return gamectx.WithGameID(userCtx, testGameID)
}

func testMetrics(t *testing.T) (*metrics.InfraMetrics, *gamemetrics.GameMetrics) {
	t.Helper()

	meter := noop.Meter{}

	infraResult, err := metrics.NewInfraMetrics(meter)
	require.NoError(t, err)

	gameResult, err := gamemetrics.NewGameMetrics(meter)
	require.NoError(t, err)

	return infraResult, gameResult
}

// setupTransaction configures mock querier to support InTransactionWithIsolation.
// It mocks BeginTx to return a mock transaction, WithTx to return a second querier (txQuerier),
// and the transaction's Commit to succeed. Returns the txQuerier for setting up in-transaction mocks.
func setupTransaction(
	t *testing.T,
	querier *mockdb.Querier,
) *mockdb.Querier {
	t.Helper()

	transaction := mockdb.NewTransaction(t)
	txQuerier := mockdb.NewQuerier(t)

	querier.EXPECT().
		BeginTx(mock.Anything, pgx.TxOptions{IsoLevel: pgx.RepeatableRead}).
		Return(transaction, nil)

	querier.EXPECT().
		WithTx(transaction).
		Return(txQuerier)

	transaction.EXPECT().
		Commit(mock.Anything).
		Return(nil)

	return txQuerier
}

// setupHappyPath configures mocks for a successful move through the entire orchestration pipeline.
// Parameters control whether a phase transition and/or game completion occurs.
func setupHappyPath[T, R any](
	t *testing.T,
	txQuerier *mockdb.Querier,
	gameStateSvc *mockstate.Service,
	validationSvc *mockorchestration.ValidationService,
	svc *mockservice.Service[T, R],
	loggingSvc *mockorchestration.LoggingService,
	missionSvc *mockmission.Service,
	move T,
	performResult R,
	moveLog sqlc.GameMoveLog,
	currentPhase sqlc.GamePhaseType,
	targetPhase sqlc.GamePhaseType,
	missionAccomplished bool,
) {
	t.Helper()

	// GetGameStateWithQuerier
	gameStateSvc.EXPECT().
		GetGameStateWithQuerier(mock.Anything, txQuerier).
		Return(&state.Game{
			ID:    testGameID,
			Turn:  testTurn,
			Phase: currentPhase,
		}, nil)

	// PhaseType — called multiple times during orchestration
	svc.EXPECT().
		PhaseType().
		Return(currentPhase)

	// Validate
	validationSvc.EXPECT().
		Validate(mock.Anything, txQuerier, mock.Anything).
		Return(nil)

	// Perform
	svc.EXPECT().
		Perform(mock.Anything, txQuerier, move).
		Return(performResult, nil)

	// LogMove
	loggingSvc.EXPECT().
		LogMove(mock.Anything, txQuerier, mock.Anything, mock.Anything).
		Return(moveLog, nil)

	// IsMissionAccomplished
	missionSvc.EXPECT().
		IsMissionAccomplished(mock.Anything, txQuerier).
		Return(missionAccomplished, nil)

	if !missionAccomplished {
		// Walk
		svc.EXPECT().
			Walk(mock.Anything, txQuerier, false).
			Return(targetPhase, nil)

		// Advance (only if phase changes)
		if targetPhase != currentPhase {
			svc.EXPECT().
				Advance(mock.Anything, txQuerier, targetPhase, performResult).
				Return(nil)
		}
	}
}

func TestOrchestrateMove_DeployEmitsMoveExecuted(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}
	moveLog := sqlc.GameMoveLog{
		ID:       99,
		GameID:   testGameID,
		PlayerID: 7,
		Phase:    sqlc.GamePhaseTypeDEPLOY,
	}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	// Verify events
	allEvents := bus.Events()
	require.Len(t, allEvents, 1, "deploy with no phase change should emit exactly 1 event")

	moveEvents := eventbus.EventsOfType[*gameevt.MoveExecuted](bus)
	require.Len(t, moveEvents, 1)
	require.Equal(t, sqlc.GamePhaseTypeDEPLOY, moveEvents[0].ActionType)
	require.Equal(t, sqlc.GamePhaseTypeDEPLOY, moveEvents[0].TargetPhase)
	require.False(t, moveEvents[0].GameOver)
	require.Equal(t, testTurn, moveEvents[0].Turn)
	require.Equal(t, moveLog.ID, moveEvents[0].MoveLog.ID)
	require.Nil(t, moveEvents[0].AttackResult)
	require.Nil(t, moveEvents[0].CardsResult)
}

func TestOrchestrateMove_AttackEmitsMoveExecutedWithResult(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[attack.Move, *attack.MoveResult](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()

	txQuerier := setupTransaction(t, querier)

	move := attack.Move{
		AttackingRegionID: "brazil",
		DefendingRegionID: "argentina",
		TroopsInSource:    5,
		TroopsInTarget:    2,
		AttackingTroops:   3,
	}
	attackResult := &attack.MoveResult{
		AttackingRegionID: "brazil",
		DefendingRegionID: "argentina",
		ConqueringTroops:  3,
	}
	moveLog := sqlc.GameMoveLog{
		ID:       100,
		GameID:   testGameID,
		PlayerID: 7,
		Phase:    sqlc.GamePhaseTypeATTACK,
	}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc,
		move, attackResult, moveLog,
		sqlc.GamePhaseTypeATTACK, sqlc.GamePhaseTypeATTACK, false,
	)

	orch := orchestration.NewOrchestrator[attack.Move, *attack.MoveResult](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	moveEvents := eventbus.EventsOfType[*gameevt.MoveExecuted](bus)
	require.Len(t, moveEvents, 1)
	require.Equal(t, sqlc.GamePhaseTypeATTACK, moveEvents[0].ActionType)
	require.NotNil(t, moveEvents[0].AttackResult)
	require.Equal(t, "brazil", moveEvents[0].AttackResult.AttackingRegionID)
	require.Equal(t, "argentina", moveEvents[0].AttackResult.DefendingRegionID)
	require.Equal(t, int64(3), moveEvents[0].AttackResult.ConqueringTroops)
	require.Nil(t, moveEvents[0].CardsResult)
}

func TestOrchestrateMove_PhaseTransitionEmitsBothEvents(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}
	moveLog := sqlc.GameMoveLog{
		ID:       101,
		GameID:   testGameID,
		PlayerID: 7,
		Phase:    sqlc.GamePhaseTypeDEPLOY,
	}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeATTACK, false,
	)

	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	allEvents := bus.Events()
	require.Len(t, allEvents, 2, "phase transition should emit MoveExecuted + PhaseTransitioned")

	moveEvents := eventbus.EventsOfType[*gameevt.MoveExecuted](bus)
	require.Len(t, moveEvents, 1)
	require.Equal(t, sqlc.GamePhaseTypeDEPLOY, moveEvents[0].ActionType)
	require.Equal(t, sqlc.GamePhaseTypeATTACK, moveEvents[0].TargetPhase)
	require.False(t, moveEvents[0].GameOver)

	phaseEvents := eventbus.EventsOfType[*gameevt.PhaseTransitioned](bus)
	require.Len(t, phaseEvents, 1)
	require.Equal(t, sqlc.GamePhaseTypeDEPLOY, phaseEvents[0].FromPhase)
	require.Equal(t, sqlc.GamePhaseTypeATTACK, phaseEvents[0].ToPhase)
	require.Equal(t, testTurn, phaseEvents[0].Turn)

	// Verify causal order: MoveExecuted before PhaseTransitioned
	require.IsType(t, &gameevt.MoveExecuted{}, allEvents[0])
	require.IsType(t, &gameevt.PhaseTransitioned{}, allEvents[1])
}

func TestOrchestrateMove_GameCompletionEmitsAllEvents(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[reinforce.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()

	txQuerier := setupTransaction(t, querier)

	move := reinforce.Move{
		SourceRegionID: "brazil",
		TargetRegionID: "argentina",
		TroopsInSource: 5,
		TroopsInTarget: 2,
		MovingTroops:   3,
	}
	moveLog := sqlc.GameMoveLog{
		ID:       102,
		GameID:   testGameID,
		PlayerID: 7,
		Phase:    sqlc.GamePhaseTypeREINFORCE,
	}

	// For game completion: mission accomplished = true, targetPhase = current phase
	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeREINFORCE, sqlc.GamePhaseTypeREINFORCE, true,
	)

	orch := orchestration.NewOrchestrator[reinforce.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	allEvents := bus.Events()
	require.Len(t, allEvents, 2, "game completion should emit MoveExecuted + GameCompleted")

	moveEvents := eventbus.EventsOfType[*gameevt.MoveExecuted](bus)
	require.Len(t, moveEvents, 1)
	require.True(t, moveEvents[0].GameOver)
	require.Equal(t, sqlc.GamePhaseTypeREINFORCE, moveEvents[0].TargetPhase)

	completedEvents := eventbus.EventsOfType[*gameevt.GameCompleted](bus)
	require.Len(t, completedEvents, 1)
	require.Equal(t, testGameID, completedEvents[0].GameID())
	require.Equal(t, testTurn, completedEvents[0].Turn)

	// Verify causal order: MoveExecuted before GameCompleted
	require.IsType(t, &gameevt.MoveExecuted{}, allEvents[0])
	require.IsType(t, &gameevt.GameCompleted{}, allEvents[1])
}

func TestOrchestrateMove_ErrorDoesNotEmit(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()

	// Setup failing transaction: BeginTx returns error
	querier.EXPECT().
		BeginTx(mock.Anything, pgx.TxOptions{IsoLevel: pgx.RepeatableRead}).
		Return(nil, errors.New("connection refused"))

	svc.EXPECT().
		PhaseType().
		Return(sqlc.GamePhaseTypeDEPLOY)

	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming,
	)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}

	err := orch.OrchestrateMove(ctx, move)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connection refused")

	allEvents := bus.Events()
	require.Empty(t, allEvents, "failed transaction should emit zero events")
}

func TestOrchestrateMove_CardsEmitsWithCardsResult(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[cards.Move, *cards.MoveResult](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()

	txQuerier := setupTransaction(t, querier)

	move := cards.Move{Combinations: []cards.CardCombination{{CardIDs: []int64{1, 2, 3}}}}
	cardsResult := &cards.MoveResult{
		ExtraDeployableTroops: 6,
		RegionTroopGrants: []cards.RegionTroopGrant{
			{RegionID: 1, RegionExternalReference: "brazil"},
		},
	}
	moveLog := sqlc.GameMoveLog{
		ID:       103,
		GameID:   testGameID,
		PlayerID: 7,
		Phase:    sqlc.GamePhaseTypeCARDS,
	}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc,
		move, cardsResult, moveLog,
		sqlc.GamePhaseTypeCARDS, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := orchestration.NewOrchestrator[cards.Move, *cards.MoveResult](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	moveEvents := eventbus.EventsOfType[*gameevt.MoveExecuted](bus)
	require.Len(t, moveEvents, 1)
	require.Equal(t, sqlc.GamePhaseTypeCARDS, moveEvents[0].ActionType)
	require.Nil(t, moveEvents[0].AttackResult)
	require.NotNil(t, moveEvents[0].CardsResult)
	require.Equal(t, int64(6), moveEvents[0].CardsResult.ExtraDeployableTroops)
}

func TestExtractResults_AllPhases(t *testing.T) {
	t.Parallel()

	// Test extractResults indirectly by verifying that each phase type produces
	// the correct result fields in the MoveExecuted event. This covers all 5 phase types.

	tests := []struct {
		name         string
		setupAndRun  func(t *testing.T) *eventbus.TestBus
		expectAttack bool
		expectCards  bool
	}{
		{
			name: "DEPLOY produces nil results",
			setupAndRun: func(t *testing.T) *eventbus.TestBus {
				t.Helper()

				return runOrchestration[deploy.Move, struct{}](
					t, sqlc.GamePhaseTypeDEPLOY,
					deploy.Move{RegionID: "r", CurrentTroops: 1, DesiredTroops: 2},
					struct{}{},
				)
			},
			expectAttack: false,
			expectCards:  false,
		},
		{
			name: "ATTACK produces AttackResult",
			setupAndRun: func(t *testing.T) *eventbus.TestBus {
				t.Helper()

				return runOrchestration[attack.Move, *attack.MoveResult](
					t,
					sqlc.GamePhaseTypeATTACK,
					attack.Move{AttackingRegionID: "a", DefendingRegionID: "b"},
					&attack.MoveResult{
						AttackingRegionID: "a",
						DefendingRegionID: "b",
						ConqueringTroops:  1,
					},
				)
			},
			expectAttack: true,
			expectCards:  false,
		},
		{
			name: "CONQUER produces nil results",
			setupAndRun: func(t *testing.T) *eventbus.TestBus {
				t.Helper()

				return runOrchestration[conquer.Move, struct{}](
					t, sqlc.GamePhaseTypeCONQUER,
					conquer.Move{Troops: 3},
					struct{}{},
				)
			},
			expectAttack: false,
			expectCards:  false,
		},
		{
			name: "REINFORCE produces nil results",
			setupAndRun: func(t *testing.T) *eventbus.TestBus {
				t.Helper()

				return runOrchestration[reinforce.Move, struct{}](
					t,
					sqlc.GamePhaseTypeREINFORCE,
					reinforce.Move{
						SourceRegionID: "a",
						TargetRegionID: "b",
						TroopsInSource: 5,
						TroopsInTarget: 1,
						MovingTroops:   2,
					},
					struct{}{},
				)
			},
			expectAttack: false,
			expectCards:  false,
		},
		{
			name: "CARDS produces CardsResult",
			setupAndRun: func(t *testing.T) *eventbus.TestBus {
				t.Helper()

				return runOrchestration[cards.Move, *cards.MoveResult](
					t, sqlc.GamePhaseTypeCARDS,
					cards.Move{Combinations: []cards.CardCombination{{CardIDs: []int64{1, 2, 3}}}},
					&cards.MoveResult{ExtraDeployableTroops: 4},
				)
			},
			expectAttack: false,
			expectCards:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			bus := test.setupAndRun(t)
			moveEvents := eventbus.EventsOfType[*gameevt.MoveExecuted](bus)
			require.Len(t, moveEvents, 1)

			if test.expectAttack {
				require.NotNil(t, moveEvents[0].AttackResult)
			} else {
				require.Nil(t, moveEvents[0].AttackResult)
			}

			if test.expectCards {
				require.NotNil(t, moveEvents[0].CardsResult)
			} else {
				require.Nil(t, moveEvents[0].CardsResult)
			}
		})
	}
}

// runOrchestration is a generic helper that runs a complete OrchestrateMove flow
// and returns the TestBus for event inspection. Used by TestExtractResults_AllPhases.
func runOrchestration[T, R any](
	t *testing.T,
	phase sqlc.GamePhaseType,
	move T,
	result R,
) *eventbus.TestBus {
	t.Helper()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[T, R](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()

	txQuerier := setupTransaction(t, querier)

	moveLog := sqlc.GameMoveLog{ID: 200, GameID: testGameID, PlayerID: 7, Phase: phase}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc,
		move, result, moveLog,
		phase, phase, false,
	)

	orch := orchestration.NewOrchestrator[T, R](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	return bus
}
