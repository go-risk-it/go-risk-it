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
	mockphase "github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/phase"
	mockstate "github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/state"
	mocksnapshot "github.com/go-risk-it/go-risk-it/internal/game/testmocks/snapshot"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
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

func (s *testStateStore) Store(gameID int64, st *apisnapshot.CachedGameState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stored[gameID] = st
}

func (s *testStateStore) Remove(gameID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.stored, gameID)
}

// testPublicSnapshot returns a fixed public snapshot for test assertions.
func testPublicSnapshot(phase apisnapshot.PhaseType) *apisnapshot.GameSnapshot {
	return &apisnapshot.GameSnapshot{
		Game: apisnapshot.GameMeta{
			ID:   testGameID,
			Turn: testTurn,
		},
		Phase: apisnapshot.Phase{
			Type:  phase,
			State: apisnapshot.EmptyPhaseState{},
		},
		Regions: []apisnapshot.RegionState{
			{ID: "brazil", OwnerID: testUserID, Troops: 5},
			{ID: "argentina", OwnerID: "player2", Troops: 3},
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

func testCachedState() *apisnapshot.CachedGameState {
	return testCachedStateForPhase(apisnapshot.PhaseDeploy)
}

func testCachedStateForPhase(phase apisnapshot.PhaseType) *apisnapshot.CachedGameState {
	return &apisnapshot.CachedGameState{
		Turn:             testTurn,
		ConqueredInTurn:  false,
		PublicSnapshot:   testPublicSnapshot(phase),
		PrivateSnapshots: testPrivateSnapshots(),
	}
}

// testReaderFactory creates a ReaderFactory for cache-miss tests.
func testReaderFactory(
	t *testing.T,
	phase apisnapshot.PhaseType,
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
			Return(testPublicSnapshot(phase), nil)
		reader.EXPECT().
			GetAllPrivateSnapshots(mock.Anything).
			Return(testPrivateSnapshots(), nil)

		return reader
	}
}

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

// setupTransaction configures a mock querier for the ReadCommitted TX that Persist opens.
func setupTransaction(
	t *testing.T,
	querier *mockdb.Querier,
) *mockdb.Querier {
	t.Helper()

	transaction := mockdb.NewTransaction(t)
	txQuerier := mockdb.NewQuerier(t)

	querier.EXPECT().
		BeginTx(mock.Anything, mock.Anything).
		Return(transaction, nil)

	querier.EXPECT().
		WithTx(transaction).
		Return(txQuerier)

	transaction.EXPECT().
		Commit(mock.Anything).
		Return(nil)

	return txQuerier
}

func sqlcPhaseToSnapshot(phase sqlc.GamePhaseType) apisnapshot.PhaseType {
	switch phase {
	case sqlc.GamePhaseTypeCARDS:
		return apisnapshot.PhaseCards
	case sqlc.GamePhaseTypeDEPLOY:
		return apisnapshot.PhaseDeploy
	case sqlc.GamePhaseTypeATTACK:
		return apisnapshot.PhaseAttack
	case sqlc.GamePhaseTypeCONQUER:
		return apisnapshot.PhaseConquer
	case sqlc.GamePhaseTypeREINFORCE:
		return apisnapshot.PhaseReinforce
	default:
		panic("unknown sqlc phase: " + string(phase))
	}
}

// setupHappyPath configures mocks for a successful move through the effects-first pipeline.
// No LoggingService, no TX-scoped Perform. Persist is handled through the TX mock.
func setupHappyPath[T, R any](
	t *testing.T,
	txQuerier *mockdb.Querier,
	validationSvc *mockorchestration.ValidationService,
	svc *mockservice.Service[T, R],
	missionSvc *mockmission.Service,
	boardSvc *mockboard.Service,
	phaseSvc *mockphase.Service,
	move T,
	performResult R,
	currentPhase sqlc.GamePhaseType,
	targetPhase sqlc.GamePhaseType,
	missionAccomplished bool,
) {
	t.Helper()

	// PhaseType — called multiple times during orchestration
	svc.EXPECT().
		PhaseType().
		Return(currentPhase)

	// Validate
	validationSvc.EXPECT().
		Validate(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// Perform (pure — no querier)
	svc.EXPECT().
		Perform(mock.Anything, move, mock.Anything).
		Return(performResult, moveservice.MoveEffect{}, nil)

	// IsMissionAccomplished — boardService.GetContinents first
	boardSvc.EXPECT().
		GetContinents(mock.Anything).
		Return(testContinents(t), nil).Maybe()

	missionSvc.EXPECT().
		IsMissionAccomplished(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(missionAccomplished, nil)

	if !missionAccomplished {
		// Walk
		svc.EXPECT().
			Walk(mock.Anything).
			Return(targetPhase, nil)

		// Advance (only if phase changes, pure — no querier)
		if targetPhase != currentPhase {
			svc.EXPECT().
				Advance(mock.Anything, targetPhase, performResult, mock.Anything).
				Return(moveservice.AdvanceEffect{}, nil)
		}
	}

	// Persist TX will use txQuerier for CreateMoveLog at minimum.
	// The empty MoveEffect means only MoveLog is written.
	txQuerier.EXPECT().
		CreateMoveLog(mock.Anything, mock.Anything).
		Return(sqlc.GameMoveLog{}, nil).Maybe()

	// Phase transition writes (when phase changes)
	if targetPhase != currentPhase && !missionAccomplished {
		phaseSvc.EXPECT().
			InsertPhase(mock.Anything, txQuerier, mock.Anything).
			Return(&sqlc.GamePhase{ID: 999}, nil).Maybe()
	}
}

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

func buildOrchestrator[T, R any](
	t *testing.T,
	querier *mockdb.Querier,
	svc *mockservice.Service[T, R],
	phaseSvc *mockphase.Service,
	gameStateSvc *mockstate.Service,
	missionSvc *mockmission.Service,
	validationSvc *mockorchestration.ValidationService,
	bus *eventbus.TestBus,
	infraM *metrics.StateMetrics,
	gameM *gamemetrics.GameMetrics,
	gameTiming *gamemetrics.GameTiming,
	readerFactory snapshot.ReaderFactory,
	store gameapi.StateStore,
	boardSvc *mockboard.Service,
	gameLocks *orchestration.GameLocks,
) orchestration.Orchestrator[T, R] {
	t.Helper()

	return orchestration.NewOrchestrator[T, R](
		querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, readerFactory, store, boardSvc, gameLocks,
	)
}

func TestOrchestrateMove_DeployEmitsMoveCompleted(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, struct{}{},
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

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
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedStateForPhase(apisnapshot.PhaseAttack))

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

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, attackResult,
		sqlc.GamePhaseTypeATTACK, sqlc.GamePhaseTypeATTACK, false,
	)

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
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
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, struct{}{},
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeATTACK, false,
	)

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
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
}

func TestOrchestrateMove_GameCompletionEmitsAllEvents(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[reinforce.Move, struct{}](t)
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedStateForPhase(apisnapshot.PhaseReinforce))

	txQuerier := setupTransaction(t, querier)

	move := reinforce.Move{
		SourceRegionID: "brazil",
		TargetRegionID: "argentina",
		TroopsInSource: 5,
		TroopsInTarget: 2,
		MovingTroops:   3,
	}

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, struct{}{},
		sqlc.GamePhaseTypeREINFORCE, sqlc.GamePhaseTypeREINFORCE, true,
	)

	// Mission victory: GameConclusion in persist TX
	txQuerier.EXPECT().
		GetPlayerByUserId(mock.Anything, testUserID).
		Return(sqlc.GamePlayer{ID: 7, UserID: testUserID, GameID: testGameID}, nil).Maybe()
	txQuerier.EXPECT().
		AssignGameWinner(mock.Anything, mock.Anything).
		Return(nil).Maybe()

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	allEvents := bus.Events()
	require.Len(t, allEvents, 2, "game completion should emit MoveCompleted + GameCompleted")

	completedEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, completedEvents, 1)
	require.True(t, completedEvents[0].GameOver)

	gameCompletedEvents := eventbus.EventsOfType[*gameevt.GameCompleted](bus)
	require.Len(t, gameCompletedEvents, 1)
	require.Equal(t, testGameID, gameCompletedEvents[0].GameID())
}

func TestOrchestrateMove_ErrorDoesNotEmit(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	svc.EXPECT().
		PhaseType().
		Return(sqlc.GamePhaseTypeDEPLOY)

	validationSvc.EXPECT().
		Validate(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// Perform returns error
	svc.EXPECT().
		Perform(mock.Anything, mock.Anything, mock.Anything).
		Return(struct{}{}, moveservice.MoveEffect{}, errors.New("perform failed"))

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
	)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}

	err := orch.OrchestrateMove(ctx, move)
	require.Error(t, err)
	require.Contains(t, err.Error(), "perform failed")

	allEvents := bus.Events()
	require.Empty(t, allEvents, "failed move should emit zero events")
}

func TestOrchestrateMove_CardsEmitsMoveCompleted(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[cards.Move, *cards.MoveResult](t)
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedStateForPhase(apisnapshot.PhaseCards))

	txQuerier := setupTransaction(t, querier)

	move := cards.Move{Combinations: []cards.CardCombination{{CardIDs: []int64{1, 2, 3}}}}
	cardsResult := &cards.MoveResult{
		ExtraDeployableTroops: 6,
		RegionTroopGrants: []cards.RegionTroopGrant{
			{RegionID: 1, RegionExternalReference: "brazil"},
		},
	}

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, cardsResult,
		sqlc.GamePhaseTypeCARDS, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	moveEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, moveEvents, 1)
	require.Equal(t, gameapi.GamePhaseTypeCARDS, moveEvents[0].ActionType)
}

func TestExtractResults_AllPhasesEmitMoveCompleted(t *testing.T) {
	t.Parallel()

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
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedStateForPhase(sqlcPhaseToSnapshot(phase)))

	txQuerier := setupTransaction(t, querier)

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, result,
		phase, phase, false,
	)

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
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
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, struct{}{},
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
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
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, struct{}{},
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	cached := store.Get(testGameID)
	require.NotNil(t, cached, "StateStore must have cached state for game")
	require.Equal(t, testTurn, cached.Turn)
	require.NotNil(t, cached.PublicSnapshot)
	require.Equal(t, testGameID, cached.PublicSnapshot.Game.ID)
}

func TestOrchestrateMove_WarmOnMiss(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore() // empty — no pre-populated state

	var capturedQuerier db.Querier
	readerFactory := testReaderFactory(t, apisnapshot.PhaseDeploy, &capturedQuerier)

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, struct{}{},
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	// Slow path: GetGameStateWithQuerier is called on cache miss using direct querier
	gameStateSvc.EXPECT().
		GetGameStateWithQuerier(mock.Anything, querier).
		Return(&state.Game{
			ID:    testGameID,
			Turn:  testTurn,
			Phase: sqlc.GamePhaseTypeDEPLOY,
		}, nil)

	// Expect HasConqueredInTurn call during warm-on-miss
	querier.EXPECT().
		HasConqueredInTurn(mock.Anything, sqlc.HasConqueredInTurnParams{
			ID:   testGameID,
			Turn: testTurn,
		}).
		Return(false, nil)

	// Expect GetAvailableDeck call during warm-on-miss
	querier.EXPECT().
		GetAvailableDeck(mock.Anything, testGameID).
		Return([]sqlc.GetAvailableDeckRow{}, nil)

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, readerFactory, store, boardSvc, gameLocks,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	// Verify the ReaderFactory was called with the direct querier (not TX querier)
	require.NotNil(t, capturedQuerier, "ReaderFactory must be called on cache miss")
	require.Equal(t, querier, capturedQuerier,
		"ReaderFactory must receive the direct querier (no TX)")

	cached := store.Get(testGameID)
	require.NotNil(t, cached, "StateStore must have cached state after warm")
	require.Equal(t, testTurn, cached.Turn)
}

func TestOrchestrateMove_CacheHitDoesNotReadDB(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, struct{}{},
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	mcEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, mcEvents, 1, "MoveCompleted must be emitted on cache hit")
}

func TestOrchestrateMove_PrevRegionsFromPrevState(t *testing.T) {
	t.Parallel()

	ctx := testCtx()
	querier := mockdb.NewQuerier(t)
	svc := mockservice.NewService[deploy.Move, struct{}](t)
	gameLocks := orchestration.NewGameLocks()
	gameStateSvc := mockstate.NewService(t)
	missionSvc := mockmission.NewService(t)
	validationSvc := mockorchestration.NewValidationService(t)
	boardSvc := mockboard.NewService(t)
	phaseSvc := mockphase.NewService(t)
	bus := eventbus.NewTestBus()
	infraM, gameM := testMetrics(t)
	gameTiming := gamemetrics.NewGameTiming()
	store := newTestStateStore()
	store.Store(testGameID, testCachedState())

	txQuerier := setupTransaction(t, querier)

	move := deploy.Move{RegionID: "brazil", CurrentTroops: 3, DesiredTroops: 5}

	setupHappyPath(
		t, txQuerier,
		validationSvc, svc, missionSvc, boardSvc, phaseSvc,
		move, struct{}{},
		sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeDEPLOY, false,
	)

	orch := buildOrchestrator(
		t, querier, svc, phaseSvc, gameStateSvc, missionSvc, validationSvc,
		bus, infraM, gameM, gameTiming, noopReaderFactory(), store, boardSvc, gameLocks,
	)

	err := orch.OrchestrateMove(ctx, move)
	require.NoError(t, err)

	mcEvents := eventbus.EventsOfType[*gameevt.MoveCompleted](bus)
	require.Len(t, mcEvents, 1)

	prevRegions := mcEvents[0].PreviousRegions
	require.Len(t, prevRegions, 2)
	require.Equal(t, "brazil", prevRegions[0].ID)
	require.Equal(t, testUserID, prevRegions[0].OwnerID)
	require.Equal(t, int64(5), prevRegions[0].Troops)
}
