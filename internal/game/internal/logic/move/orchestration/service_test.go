package orchestration_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	apisnapshot "github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	gamemetrics "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/metrics"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/reinforce"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/snapshot"
	mockdb "github.com/go-risk-it/go-risk-it/internal/game/testmocks/data/db"
	mockboard "github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/board"
	mockmission "github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/mission"
	mockorchestration "github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/move/orchestration"
	mockservice "github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/move/service"
	mockstate "github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/state"
	mocksnapshot "github.com/go-risk-it/go-risk-it/internal/game/testmocks/snapshot"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
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

// testStateStore is a simple spy that records Store calls for verification.
type testStateStore struct {
	mu     sync.Mutex
	stored map[int64]*apisnapshot.CachedGameState
}

func newTestStateStore() *testStateStore {
	return &testStateStore{
		stored: make(map[int64]*apisnapshot.CachedGameState),
	}
}

func (s *testStateStore) Get(gameID int64) *apisnapshot.CachedGameState {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.stored[gameID]
}

func (s *testStateStore) Store(gameID int64, state *apisnapshot.CachedGameState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stored[gameID] = state
}

func (s *testStateStore) Remove(gameID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.stored, gameID)
}

// testPublicSnapshot returns a fixed public snapshot for test assertions.
func testPublicSnapshot() *apisnapshot.GameSnapshot {
	return &apisnapshot.GameSnapshot{
		Game: apisnapshot.GameMeta{
			ID:   testGameID,
			Turn: testTurn,
		},
		Phase: apisnapshot.Phase{
			Type:  apisnapshot.PhaseDeploy,
			State: apisnapshot.EmptyPhaseState{},
		},
		Regions: []apisnapshot.RegionState{
			{ID: "brazil", OwnerID: testUserID, Troops: 5},
		},
		Players: []apisnapshot.PlayerState{
			{
				UserID:    testUserID,
				Name:      "Player 1",
				Index:     0,
				CardCount: 2,
				Status:    apisnapshot.PlayerAlive,
			},
		},
	}
}

// testPrivateSnapshots returns a fixed private snapshots map for test assertions.
func testPrivateSnapshots() map[string]*apisnapshot.PlayerPrivate {
	return map[string]*apisnapshot.PlayerPrivate{
		testUserID: {
			Cards: []apisnapshot.CardState{
				{ID: 1, Type: apisnapshot.CardInfantry, Region: "brazil"},
			},
			Mission: apisnapshot.PlayerMission{
				Type:   apisnapshot.MissionTwentyFourTerritories,
				Detail: apisnapshot.TwentyFourTerritoriesMission{},
			},
		},
	}
}

// testCachedState returns a fixed CachedGameState for pre-populating the store.
func testCachedState() *apisnapshot.CachedGameState {
	return &apisnapshot.CachedGameState{
		Turn:             testTurn,
		ConqueredInTurn:  false,
		PublicSnapshot:   testPublicSnapshot(),
		PrivateSnapshots: testPrivateSnapshots(),
	}
}

// testReaderFactory creates a ReaderFactory that returns a mock Reader
// configured to return the test snapshots. Used for warm-on-miss tests.
func testReaderFactory(
	t *testing.T,
	capturedQuerier *db.Querier,
) snapshot.ReaderFactory {
	t.Helper()

	return func(querier db.Querier) gameapi.SnapshotReader {
		if capturedQuerier != nil {
			*capturedQuerier = querier
		}

		reader := mocksnapshot.NewReader(t)
		reader.EXPECT().
			GetPublicSnapshot(mock.Anything).
			Return(testPublicSnapshot(), nil)
		reader.EXPECT().
			GetAllPrivateSnapshots(mock.Anything).
			Return(testPrivateSnapshots(), nil)

		return reader
	}
}

// noopReaderFactory returns a ReaderFactory that panics if called.
// Used for cache-hit tests where the factory should never be invoked.
func noopReaderFactory() snapshot.ReaderFactory {
	return func(_ db.Querier) gameapi.SnapshotReader {
		panic("ReaderFactory should not be called on cache hit")
	}
}

func testCtx() gamectx.GameContext {
	traceCtx := kernelctx.WithSpan(context.Background(), tracenoop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, testUserID)

	return gamectx.WithGameID(userCtx, testGameID)
}

func testMetrics(t *testing.T) (*metrics.StateMetrics, *gamemetrics.GameMetrics) {
	t.Helper()

	meter := noop.Meter{}

	stateResult, err := metrics.NewStateMetrics(meter)
	require.NoError(t, err)

	gameResult, err := gamemetrics.NewGameMetrics(meter)
	require.NoError(t, err)

	return stateResult, gameResult
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
// When phase transitions occur (targetPhase != currentPhase), boardSvc must be non-nil.
func setupHappyPath[T, R any](
	t *testing.T,
	txQuerier *mockdb.Querier,
	gameStateSvc *mockstate.Service,
	validationSvc *mockorchestration.ValidationService,
	svc *mockservice.Service[T, R],
	loggingSvc *mockorchestration.LoggingService,
	missionSvc *mockmission.Service,
	boardSvc *mockboard.Service,
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
		Perform(mock.Anything, txQuerier, move, mock.Anything).
		Return(performResult, moveservice.MoveEffect{}, nil)

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
			Walk(mock.Anything).
			Return(targetPhase, nil)

		// Advance (only if phase changes)
		if targetPhase != currentPhase {
			// buildAdvanceContext calls boardService.GetContinents
			boardSvc.EXPECT().
				GetContinents(mock.Anything).
				Return(testContinents(t), nil)

			svc.EXPECT().
				Advance(mock.Anything, txQuerier, targetPhase, performResult, mock.Anything).
				Return(moveservice.AdvanceEffect{}, nil)
		}
	}
}

// testContinents builds a minimal board.Continents for test use.
func testContinents(t *testing.T) board.Continents {
	t.Helper()

	continents, err := board.NewContinents(&board.BoardDto{
		Regions: []board.RegionDto{
			{ExternalReference: "brazil", Continent: "south_america"},
		},
		Continents: []board.ContinentDto{
			{ExternalReference: "south_america", BonusTroops: 2},
		},
	})
	require.NoError(t, err)

	return continents
}

func TestOrchestrateMove_DeployEmitsMoveCompleted(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

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
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	// Verify MoveCompleted event
	moveEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, moveEvents, 1)
	require.Equal(t, gameapi.GamePhaseTypeDEPLOY, moveEvents[0].ActionType)
	require.Equal(t, gameapi.GamePhaseTypeDEPLOY, moveEvents[0].TargetPhase)
	require.False(t, moveEvents[0].GameOver)
	require.Equal(t, testTurn, moveEvents[0].Turn)
}

func TestOrchestrateMove_AttackEmitsMoveCompleted(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[attack.Move, *attack.MoveResult](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

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
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, attackResult, moveLog,
		sqlc.GamePhaseTypeATTACK, sqlc.GamePhaseTypeATTACK, false,
	)

	orch := orchestration.NewOrchestrator[attack.Move, *attack.MoveResult](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	moveEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, moveEvents, 1)
	require.Equal(t, gameapi.GamePhaseTypeATTACK, moveEvents[0].ActionType)
}

func TestOrchestrateMove_PhaseTransitionInMoveCompleted(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

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
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeATTACK, false,
	)

	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	allEvents := bus.Events()
	require.Len(t, allEvents, 1, "phase transition should emit single MoveCompleted")

	moveEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, moveEvents, 1)
	require.Equal(t, gameapi.GamePhaseTypeDEPLOY, moveEvents[0].ActionType)
	require.Equal(t, gameapi.GamePhaseTypeATTACK, moveEvents[0].TargetPhase)
	require.False(t, moveEvents[0].GameOver)
	require.Equal(t, testTurn, moveEvents[0].Turn)
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
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

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
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeREINFORCE, sqlc.GamePhaseTypeREINFORCE, true,
	)

	orch := orchestration.NewOrchestrator[reinforce.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	allEvents := bus.Events()
	require.Len(t, allEvents, 1, "game completion should emit single MoveCompleted")

	completedEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, completedEvents, 1)
	require.True(t, completedEvents[0].GameOver)
	require.Equal(t, gameapi.GamePhaseTypeREINFORCE, completedEvents[0].TargetPhase)
	require.Equal(t, testGameID, completedEvents[0].GameID())
	require.Equal(t, testTurn, completedEvents[0].Turn)
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
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	// Setup failing transaction: BeginTx returns error
	querier.EXPECT().
		BeginTx(mock.Anything, pgx.TxOptions{IsoLevel: pgx.RepeatableRead}).
		Return(nil, errors.New("connection refused"))

	svc.EXPECT().
		PhaseType().
		Return(sqlc.GamePhaseTypeDEPLOY)

	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}

	err := orch.OrchestrateMove(ctx, move)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connection refused")

	allEvents := bus.Events()
	require.Empty(t, allEvents, "failed transaction should emit zero events")
}

func TestOrchestrateMove_CardsEmitsMoveCompleted(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[cards.Move, *cards.MoveResult](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

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
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, cardsResult, moveLog,
		sqlc.GamePhaseTypeCARDS, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := orchestration.NewOrchestrator[cards.Move, *cards.MoveResult](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	moveEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, moveEvents, 1)
	require.Equal(t, gameapi.GamePhaseTypeCARDS, moveEvents[0].ActionType)
}

func TestExtractResults_AllPhasesEmitMoveCompleted(t *testing.T) {
	t.Parallel()

	// Verify that each phase type emits a MoveCompleted event.

	tests := []struct {
		name        string
		setupAndRun func(t *testing.T) *eventbus.TestBus
	}{
		{
			name: "DEPLOY emits MoveCompleted",
			setupAndRun: func(t *testing.T) *eventbus.TestBus {
				t.Helper()

				return runOrchestration[deploy.Move, struct{}](
					t, sqlc.GamePhaseTypeDEPLOY,
					deploy.Move{RegionID: "r", CurrentTroops: 1, DesiredTroops: 2},
					struct{}{},
				)
			},
		},
		{
			name: "ATTACK emits MoveCompleted",
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
		},
		{
			name: "CONQUER emits MoveCompleted",
			setupAndRun: func(t *testing.T) *eventbus.TestBus {
				t.Helper()

				return runOrchestration[conquer.Move, struct{}](
					t, sqlc.GamePhaseTypeCONQUER,
					conquer.Move{Troops: 3},
					struct{}{},
				)
			},
		},
		{
			name: "REINFORCE emits MoveCompleted",
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
		},
		{
			name: "CARDS emits MoveCompleted",
			setupAndRun: func(t *testing.T) *eventbus.TestBus {
				t.Helper()

				return runOrchestration[cards.Move, *cards.MoveResult](
					t, sqlc.GamePhaseTypeCARDS,
					cards.Move{Combinations: []cards.CardCombination{{CardIDs: []int64{1, 2, 3}}}},
					&cards.MoveResult{ExtraDeployableTroops: 4},
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			bus := test.setupAndRun(t)
			moveEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
			require.Len(t, moveEvents, 1)
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
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	moveLog := sqlc.GameMoveLog{ID: 200, GameID: testGameID, PlayerID: 7, Phase: phase}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, result, moveLog,
		phase, phase, false,
	)

	orch := orchestration.NewOrchestrator[T, R](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	return bus
}

// ---------------------------------------------------------------------------
// ECST-specific tests
// ---------------------------------------------------------------------------

func TestOrchestrateMove_EmitsMoveCompleted(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}
	moveLog := sqlc.GameMoveLog{
		ID:       110,
		GameID:   testGameID,
		PlayerID: 7,
		Phase:    sqlc.GamePhaseTypeDEPLOY,
	}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	mcEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, mcEvents, 1, "MoveCompleted must be emitted")

	moveCompleted := mcEvents[0]
	require.Equal(t, testGameID, moveCompleted.GameID())
	require.Equal(t, gameapi.GamePhaseTypeDEPLOY, moveCompleted.ActionType)
	require.Equal(t, gameapi.GamePhaseTypeDEPLOY, moveCompleted.TargetPhase)
	require.Equal(t, testTurn, moveCompleted.Turn)
	require.False(t, moveCompleted.GameOver)
	require.NotNil(t, moveCompleted.PublicSnapshot)
	require.Equal(t, testGameID, moveCompleted.PublicSnapshot.Game.ID)
	require.NotNil(t, moveCompleted.PrivateSnapshots)
	require.Contains(t, moveCompleted.PrivateSnapshots, testUserID)
}

func TestOrchestrateMove_StoresStateAfterCommit(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}
	moveLog := sqlc.GameMoveLog{
		ID:       111,
		GameID:   testGameID,
		PlayerID: 7,
		Phase:    sqlc.GamePhaseTypeDEPLOY,
	}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	// Verify StateStore.Store was called with the BuildNewState output
	cached := store.Get(testGameID)
	require.NotNil(t, cached, "StateStore must have cached state for game")
	require.Equal(t, testTurn, cached.Turn)
	require.NotNil(t, cached.PublicSnapshot)
	require.Equal(t, testGameID, cached.PublicSnapshot.Game.ID)
	require.NotNil(t, cached.PrivateSnapshots)
	require.Contains(t, cached.PrivateSnapshots, testUserID)
}

// TestOrchestrateMove_WarmOnMiss verifies that when the state store has no
// cached state (cache miss), the pipeline warms from the database using the
// transactional querier and stores the result.
func TestOrchestrateMove_WarmOnMiss(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore() // empty — no pre-populated state

	// Capture the querier passed to the ReaderFactory
	var capturedQuerier db.Querier
	readerFactory := testReaderFactory(t, &capturedQuerier)

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}
	moveLog := sqlc.GameMoveLog{
		ID:       112,
		GameID:   testGameID,
		PlayerID: 7,
		Phase:    sqlc.GamePhaseTypeDEPLOY,
	}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	// Expect HasConqueredInTurn call during warm-on-miss
	txQuerier.EXPECT().
		HasConqueredInTurn(mock.Anything, sqlc.HasConqueredInTurnParams{
			ID:   testGameID,
			Turn: testTurn,
		}).
		Return(false, nil)

	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, readerFactory, store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	// Verify the ReaderFactory was called with the transactional querier
	require.NotNil(t, capturedQuerier, "ReaderFactory must be called on cache miss")
	require.Equal(t, txQuerier, capturedQuerier,
		"ReaderFactory must receive the transactional querier")

	// Verify state was stored and MoveCompleted emitted
	cached := store.Get(testGameID)
	require.NotNil(t, cached, "StateStore must have cached state after warm")
	require.Equal(t, testTurn, cached.Turn)

	mcEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, mcEvents, 1)
	require.NotNil(t, mcEvents[0].PublicSnapshot)
	require.Contains(t, mcEvents[0].PrivateSnapshots, testUserID)
}

// TestOrchestrateMove_CacheHitDoesNotReadDB verifies that when the state store
// has cached state, the pipeline does NOT call the ReaderFactory (no DB reads).
func TestOrchestrateMove_CacheHitDoesNotReadDB(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}
	moveLog := sqlc.GameMoveLog{
		ID:       113,
		GameID:   testGameID,
		PlayerID: 7,
		Phase:    sqlc.GamePhaseTypeDEPLOY,
	}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	// Use panicking reader factory — proves it's never called on cache hit
	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	mcEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, mcEvents, 1, "MoveCompleted must be emitted on cache hit")
}

// TestOrchestrateMove_PrevRegionsFromPrevState verifies that the MoveCompleted
// event carries the regions from the pre-mutation state (for headline detection).
func TestOrchestrateMove_PrevRegionsFromPrevState(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameStateSvc := mockstate.NewService(t)
	loggingSvc := mockorchestration.NewLoggingService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}
	moveLog := sqlc.GameMoveLog{
		ID:       114,
		GameID:   testGameID,
		PlayerID: 7,
		Phase:    sqlc.GamePhaseTypeDEPLOY,
	}

	setupHappyPath(
		t, txQuerier,
		gameStateSvc, validationSvc, svc, loggingSvc, missionSvc, boardSvc,
		move, struct{}{}, moveLog,
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := orchestration.NewOrchestrator[deploy.Move, struct{}](
		querier, svc, gameStateSvc, loggingSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	mcEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, mcEvents, 1)

	// PreviousRegions should match the pre-mutation snapshot regions
	prevRegions := mcEvents[0].PreviousRegions
	require.Len(t, prevRegions, 1)
	require.Equal(t, "brazil", prevRegions[0].ID)
	require.Equal(t, testUserID, prevRegions[0].OwnerID)
	require.Equal(t, int64(5), prevRegions[0].Troops)
}
