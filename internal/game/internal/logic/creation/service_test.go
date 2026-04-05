package creation_test

import (
	"errors"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/handlers"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/creation"
	gamemetrics "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/metrics"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/card"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/mission"
	playermock "github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/region"
	snapshotmock "github.com/go-risk-it/go-risk-it/internal/game/testmocks/snapshot"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

var (
	errCreatePlayers = errors.New("error inserting players")
	errInsertGame    = errors.New("insert logic error")
)

func testStateMetrics(t *testing.T) *metrics.StateMetrics {
	t.Helper()

	m, err := metrics.NewStateMetrics(noop.Meter{})
	require.NoError(t, err)

	return m
}

func testGameMetrics(t *testing.T) *gamemetrics.GameMetrics {
	t.Helper()

	m, err := gamemetrics.NewGameMetrics(noop.Meter{})
	require.NoError(t, err)

	return m
}

// creates a game with a valid board and list of users.
func TestServiceImpl_CreateGame_WithValidBoardAndUsers(t *testing.T) {
	t.Parallel()

	gameID := int64(1)
	phaseID := int64(1)
	users := []player.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "Giovanni"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "Gabriele"},
	}
	context := kernelctx.WithUserID(
		kernelctx.WithSpan(t.Context(), tracenoop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)

	mockQuerier := db.NewQuerier(t)

	players := []sqlc.GamePlayer{
		{ID: 420, TurnIndex: 1, GameID: gameID, UserID: "Giovanni"},
		{ID: 69, TurnIndex: 2, GameID: gameID, UserID: "Gabriele"},
	}

	regions := []string{
		"netherlands",
		"italy",
		"tasin",
		"samon",
	}

	// setup mocks
	mockQuerier.EXPECT().InsertGame(mock.Anything).Return(sqlc.GameGame{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: 1, Valid: true},
	}, nil)

	mockQuerier.EXPECT().InsertPhase(mock.Anything, sqlc.InsertPhaseParams{
		GameID: gameID,
		Type:   sqlc.GamePhaseTypeDEPLOY,
		Turn:   0,
	}).Return(sqlc.GamePhase{ID: phaseID}, nil)

	mockQuerier.EXPECT().SetGamePhase(mock.Anything, sqlc.SetGamePhaseParams{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: phaseID, Valid: true},
	}).Return(nil)

	mockQuerier.EXPECT().InsertDeployPhase(mock.Anything, sqlc.InsertDeployPhaseParams{
		PhaseID:          phaseID,
		DeployableTroops: int64(3),
	}).Return(sqlc.GameDeployPhase{ID: 1}, nil)

	playerServiceMock := playermock.NewService(t)
	playerServiceMock.
		EXPECT().
		CreatePlayers(mock.Anything, mockQuerier, gameID, users).
		Return(players, nil)

	missionServiceMock := mission.NewService(t)
	missionServiceMock.
		EXPECT().
		CreateMissions(mock.Anything, mockQuerier, players).
		Return(nil)

	regionServiceMock := region.NewService(t)
	regionServiceMock.
		EXPECT().
		CreateRegions(mock.Anything, mockQuerier, players, regions).
		Return(nil)

	cardServiceMock := card.NewService(t)
	cardServiceMock.
		EXPECT().
		CreateCards(mock.Anything, mockQuerier).
		Return(nil)

	// Initialize the state
	service := creation.NewService(
		mockQuerier,
		eventbus.NewTestBus(),
		snapshotmock.NewReader(t),
		handlers.NewStateStore(),
		cardServiceMock,
		missionServiceMock,
		playerServiceMock,
		regionServiceMock,
		testStateMetrics(t),
		testGameMetrics(t),
		gamemetrics.NewGameTiming(),
	)

	gameID, err := service.CreateGameWithQuerier(context, mockQuerier, regions, users)

	require.NoError(t, err)
	require.Equal(t, int64(1), gameID)
}

// returns error if InsertGame method returns an error.
func TestServiceImpl_CreateGame_InsertGameError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	cardService := card.NewService(t)
	missionService := mission.NewService(t)
	playerService := playermock.NewService(t)
	regionService := region.NewService(t)
	querier := db.NewQuerier(t)

	// Initialize the state under test
	service := creation.NewService(
		querier,
		eventbus.NewTestBus(),
		snapshotmock.NewReader(t),
		handlers.NewStateStore(),
		cardService,
		missionService,
		playerService,
		regionService,
		testStateMetrics(t),
		testGameMetrics(t),
		gamemetrics.NewGameTiming(),
	)

	// Set up test data
	ctx := kernelctx.WithUserID(
		kernelctx.WithSpan(t.Context(), tracenoop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)
	users := []player.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}

	// Set up expectations for InsertGame method
	querier.
		EXPECT().
		InsertGame(mock.Anything).Return(sqlc.GameGame{}, errInsertGame)

	// Call the method under test
	gameID, err := service.CreateGameWithQuerier(ctx, querier, []string{}, users)

	// Assert the result
	require.Error(t, err)
	require.EqualError(t, err, "failed to insert game: insert logic error")
	require.Equal(t, int64(-1), gameID)
}

// returns error if CreatePlayers method returns an error.
func TestServiceImpl_CreateGame_CreatePlayersError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	querier := db.NewQuerier(t)
	cardService := card.NewService(t)
	missionService := mission.NewService(t)
	playerService := playermock.NewService(t)
	regionService := region.NewService(t)

	// Initialize the state under test
	service := creation.NewService(
		querier,
		eventbus.NewTestBus(),
		snapshotmock.NewReader(t),
		handlers.NewStateStore(),
		cardService,
		missionService,
		playerService,
		regionService,
		testStateMetrics(t),
		testGameMetrics(t),
		gamemetrics.NewGameTiming(),
	)

	// Set up test data
	context := kernelctx.WithUserID(
		kernelctx.WithSpan(t.Context(), tracenoop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)
	users := []player.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}
	gameID := int64(1)

	// Set up expectations for InsertGame method
	querier.
		EXPECT().
		InsertGame(mock.Anything).
		Return(sqlc.GameGame{
			ID: gameID,
		}, nil)

	// Set up expectations for CreatePlayers method
	playerService.
		EXPECT().
		CreatePlayers(mock.Anything, querier, int64(1), users).
		Return(nil, errCreatePlayers)

	// Call the method under test
	gameID, err := service.CreateGameWithQuerier(context, querier, []string{}, users)

	// Assert the result
	require.Error(t, err)
	require.EqualError(t, err, "failed to create players: error inserting players")
	require.Equal(t, int64(-1), gameID)
}

// verifies that CreateGameWithQuerier does not emit events (emission is in CreateGame wrapper).
func TestServiceImpl_CreateGameWithQuerier_NoEventEmitted(t *testing.T) {
	t.Parallel()

	gameID := int64(1)
	phaseID := int64(1)
	users := []player.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "Giovanni"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "Gabriele"},
	}
	context := kernelctx.WithUserID(
		kernelctx.WithSpan(t.Context(), tracenoop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)

	mockQuerier := db.NewQuerier(t)

	players := []sqlc.GamePlayer{
		{ID: 420, TurnIndex: 1, GameID: gameID, UserID: "Giovanni"},
		{ID: 69, TurnIndex: 2, GameID: gameID, UserID: "Gabriele"},
	}

	regions := []string{"netherlands", "italy", "tasin", "samon"}

	mockQuerier.EXPECT().InsertGame(mock.Anything).Return(sqlc.GameGame{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: 1, Valid: true},
	}, nil)

	mockQuerier.EXPECT().InsertPhase(mock.Anything, sqlc.InsertPhaseParams{
		GameID: gameID,
		Type:   sqlc.GamePhaseTypeDEPLOY,
		Turn:   0,
	}).Return(sqlc.GamePhase{ID: phaseID}, nil)

	mockQuerier.EXPECT().SetGamePhase(mock.Anything, sqlc.SetGamePhaseParams{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: phaseID, Valid: true},
	}).Return(nil)

	mockQuerier.EXPECT().InsertDeployPhase(mock.Anything, sqlc.InsertDeployPhaseParams{
		PhaseID:          phaseID,
		DeployableTroops: int64(3),
	}).Return(sqlc.GameDeployPhase{ID: 1}, nil)

	playerServiceMock := playermock.NewService(t)
	playerServiceMock.EXPECT().
		CreatePlayers(mock.Anything, mockQuerier, gameID, users).
		Return(players, nil)

	missionServiceMock := mission.NewService(t)
	missionServiceMock.EXPECT().
		CreateMissions(mock.Anything, mockQuerier, players).
		Return(nil)

	regionServiceMock := region.NewService(t)
	regionServiceMock.EXPECT().
		CreateRegions(mock.Anything, mockQuerier, players, regions).
		Return(nil)

	cardServiceMock := card.NewService(t)
	cardServiceMock.EXPECT().
		CreateCards(mock.Anything, mockQuerier).
		Return(nil)

	bus := eventbus.NewTestBus()
	service := creation.NewService(
		mockQuerier,
		bus,
		snapshotmock.NewReader(t),
		handlers.NewStateStore(),
		cardServiceMock,
		missionServiceMock,
		playerServiceMock,
		regionServiceMock,
		testStateMetrics(t),
		testGameMetrics(t),
		gamemetrics.NewGameTiming(),
	)

	result, err := service.CreateGameWithQuerier(context, mockQuerier, regions, users)

	require.NoError(t, err)
	require.Equal(t, int64(1), result)
	require.Empty(
		t,
		bus.Events(),
		"CreateGameWithQuerier must not emit events; emission is in CreateGame",
	)
}

// verifies that no events are emitted when CreateGameWithQuerier fails.
func TestServiceImpl_CreateGameWithQuerier_Error_NoEventEmitted(t *testing.T) {
	t.Parallel()

	querier := db.NewQuerier(t)
	bus := eventbus.NewTestBus()

	service := creation.NewService(
		querier,
		bus,
		snapshotmock.NewReader(t),
		handlers.NewStateStore(),
		card.NewService(t),
		mission.NewService(t),
		playermock.NewService(t),
		region.NewService(t),
		testStateMetrics(t),
		testGameMetrics(t),
		gamemetrics.NewGameTiming(),
	)

	userCtx := kernelctx.WithUserID(
		kernelctx.WithSpan(t.Context(), tracenoop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)
	users := []player.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}

	querier.EXPECT().InsertGame(mock.Anything).Return(sqlc.GameGame{}, errInsertGame)

	_, err := service.CreateGameWithQuerier(userCtx, querier, []string{}, users)

	require.Error(t, err)
	require.Empty(t, bus.Events(), "no events should be emitted when CreateGameWithQuerier fails")
}

// ---------------------------------------------------------------------------
// CreateGame integration tests (post-commit snapshot + StateStore)
// ---------------------------------------------------------------------------

// setupTransactionalQuerier configures a mock querier that supports
// BeginTx/WithTx/Commit for InTransaction, returning a txQuerier
// that handles the full game creation SQL sequence.
func setupTransactionalQuerier(
	t *testing.T,
	gameID int64,
	phaseID int64,
) (*db.Querier, *db.Querier) {
	t.Helper()

	poolQuerier := db.NewQuerier(t)
	txQuerier := db.NewQuerier(t)
	tx := db.NewTransaction(t)

	// Transaction lifecycle
	poolQuerier.EXPECT().
		BeginTx(mock.Anything, mock.Anything).
		Return(tx, nil)
	poolQuerier.EXPECT().
		WithTx(tx).
		Return(txQuerier)
	tx.EXPECT().
		Commit(mock.Anything).
		Return(nil)

	// Game creation SQL sequence on the transactional querier
	txQuerier.EXPECT().
		InsertGame(mock.Anything).
		Return(sqlc.GameGame{
			ID:             gameID,
			CurrentPhaseID: pgtype.Int8{Int64: 1, Valid: true},
		}, nil)

	txQuerier.EXPECT().
		InsertPhase(mock.Anything, sqlc.InsertPhaseParams{
			GameID: gameID,
			Type:   sqlc.GamePhaseTypeDEPLOY,
			Turn:   0,
		}).
		Return(sqlc.GamePhase{ID: phaseID}, nil)

	txQuerier.EXPECT().
		SetGamePhase(mock.Anything, sqlc.SetGamePhaseParams{
			ID:             gameID,
			CurrentPhaseID: pgtype.Int8{Int64: phaseID, Valid: true},
		}).
		Return(nil)

	txQuerier.EXPECT().
		InsertDeployPhase(mock.Anything, sqlc.InsertDeployPhaseParams{
			PhaseID:          phaseID,
			DeployableTroops: int64(3),
		}).
		Return(sqlc.GameDeployPhase{ID: 1}, nil)

	return poolQuerier, txQuerier
}

func testPublicSnapshot(gameID int64) *snapshot.GameSnapshot {
	return &snapshot.GameSnapshot{
		Game: snapshot.GameMeta{ID: gameID, Turn: 0},
		Phase: snapshot.Phase{
			Type: snapshot.PhaseDeploy,
			State: snapshot.DeployPhaseState{
				DeployableTroops: 3,
			},
		},
		Regions: []snapshot.RegionState{
			{ID: "netherlands", OwnerID: "user-1", Troops: 1},
			{ID: "italy", OwnerID: "user-2", Troops: 1},
		},
		Players: []snapshot.PlayerState{
			{UserID: "user-1", Name: "Giovanni", Index: 0, CardCount: 0, Status: snapshot.PlayerAlive},
			{UserID: "user-2", Name: "Gabriele", Index: 1, CardCount: 0, Status: snapshot.PlayerAlive},
		},
	}
}

func testPrivateSnapshots() map[string]*snapshot.PlayerPrivate {
	return map[string]*snapshot.PlayerPrivate{
		"user-1": {
			Cards: []snapshot.CardState{},
			Mission: snapshot.PlayerMission{
				Type:   snapshot.MissionTwentyFourTerritories,
				Detail: snapshot.TwentyFourTerritoriesMission{},
			},
		},
		"user-2": {
			Cards: []snapshot.CardState{},
			Mission: snapshot.PlayerMission{
				Type:   snapshot.MissionEighteenTerritoriesTwoTroops,
				Detail: snapshot.EighteenTerritoriesTwoTroopsMission{},
			},
		},
	}
}

// TestCreate_StoresInitialState verifies that CreateGame reads the initial
// snapshot post-commit and stores it in the StateStore.
func TestCreate_StoresInitialState(t *testing.T) {
	t.Parallel()

	gameID := int64(42)
	phaseID := int64(1)
	users := []player.Player{
		{UserID: "user-1", Name: "Giovanni"},
		{UserID: "user-2", Name: "Gabriele"},
	}
	regions := []string{"netherlands", "italy", "tasin", "samon"}

	poolQuerier, txQuerier := setupTransactionalQuerier(t, gameID, phaseID)

	playerServiceMock := playermock.NewService(t)
	playerServiceMock.EXPECT().
		CreatePlayers(mock.Anything, txQuerier, gameID, users).
		Return([]sqlc.GamePlayer{
			{ID: 1, TurnIndex: 0, GameID: gameID, UserID: "user-1"},
			{ID: 2, TurnIndex: 1, GameID: gameID, UserID: "user-2"},
		}, nil)

	missionServiceMock := mission.NewService(t)
	missionServiceMock.EXPECT().
		CreateMissions(mock.Anything, txQuerier, mock.Anything).
		Return(nil)

	regionServiceMock := region.NewService(t)
	regionServiceMock.EXPECT().
		CreateRegions(mock.Anything, txQuerier, mock.Anything, regions).
		Return(nil)

	cardServiceMock := card.NewService(t)
	cardServiceMock.EXPECT().
		CreateCards(mock.Anything, txQuerier).
		Return(nil)

	// Snapshot reader returns initial state
	publicSnap := testPublicSnapshot(gameID)
	privateSnaps := testPrivateSnapshots()

	readerMock := snapshotmock.NewReader(t)
	readerMock.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(publicSnap, nil)
	readerMock.EXPECT().
		GetAllPrivateSnapshots(mock.Anything).
		Return(privateSnaps, nil)

	stateStore := handlers.NewStateStore()
	bus := eventbus.NewTestBus()

	svc := creation.NewService(
		poolQuerier,
		bus,
		readerMock,
		stateStore,
		cardServiceMock,
		missionServiceMock,
		playerServiceMock,
		regionServiceMock,
		testStateMetrics(t),
		testGameMetrics(t),
		gamemetrics.NewGameTiming(),
	)

	userCtx := kernelctx.WithUserID(
		kernelctx.WithSpan(t.Context(), tracenoop.Span{}),
		"user-1",
	)

	result, err := svc.CreateGame(userCtx, 0, regions, users)

	require.NoError(t, err)
	require.Equal(t, gameID, result)

	// Verify StateStore has the cached state
	cached := stateStore.Get(gameID)
	require.NotNil(t, cached, "StateStore must contain cached state after CreateGame")
	require.Equal(t, int64(0), cached.Turn)
	require.Equal(t, publicSnap, cached.PublicSnapshot)
	require.Equal(t, privateSnaps, cached.PrivateSnapshots)

	// Verify enriched GameCreated event was emitted
	events := eventbus.EventsOfType[*gameevt.GameCreated](bus)
	require.Len(t, events, 1)
	require.Equal(t, gameID, events[0].GameID())
	require.Equal(t, 2, events[0].NumPlayers)
	require.Equal(t, publicSnap, events[0].PublicSnapshot)
	require.Equal(t, privateSnaps, events[0].PrivateSnapshots)
}

// TestGameCreated_CarriesSnapshot verifies that a GameCreated event
// constructed with snapshot data has non-nil snapshot fields.
func TestGameCreated_CarriesSnapshot(t *testing.T) {
	t.Parallel()

	publicSnap := testPublicSnapshot(42)
	privateSnaps := testPrivateSnapshots()

	event := gameevt.NewGameCreated(42, 0, time.Now(), 4, publicSnap, privateSnaps)

	require.Equal(t, gameevt.TypeGameCreated, event.EventType())
	require.Equal(t, int64(42), event.GameID())
	require.Equal(t, 4, event.NumPlayers)
	require.NotNil(t, event.PublicSnapshot)
	require.NotNil(t, event.PrivateSnapshots)
	require.Equal(t, publicSnap, event.PublicSnapshot)
	require.Equal(t, privateSnaps, event.PrivateSnapshots)

	// ToRecord must NOT contain snapshot data
	record := event.ToRecord()
	require.NotContains(t, record, "public_snapshot")
	require.NotContains(t, record, "private_snapshots")
	require.Equal(t, gameevt.TypeGameCreated, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, 4, record["num_players"])
}

// TestCreate_SnapshotReadError_EmitsLeanEvent verifies that when snapshot
// reading fails post-commit, the game is still created successfully and
// a lean GameCreated event (without snapshots) is emitted.
func TestCreate_SnapshotReadError_EmitsLeanEvent(t *testing.T) {
	t.Parallel()

	gameID := int64(42)
	phaseID := int64(1)
	users := []player.Player{
		{UserID: "user-1", Name: "Giovanni"},
		{UserID: "user-2", Name: "Gabriele"},
	}
	regions := []string{"netherlands", "italy", "tasin", "samon"}

	poolQuerier, txQuerier := setupTransactionalQuerier(t, gameID, phaseID)

	playerServiceMock := playermock.NewService(t)
	playerServiceMock.EXPECT().
		CreatePlayers(mock.Anything, txQuerier, gameID, users).
		Return([]sqlc.GamePlayer{
			{ID: 1, TurnIndex: 0, GameID: gameID, UserID: "user-1"},
			{ID: 2, TurnIndex: 1, GameID: gameID, UserID: "user-2"},
		}, nil)

	missionServiceMock := mission.NewService(t)
	missionServiceMock.EXPECT().
		CreateMissions(mock.Anything, txQuerier, mock.Anything).
		Return(nil)

	regionServiceMock := region.NewService(t)
	regionServiceMock.EXPECT().
		CreateRegions(mock.Anything, txQuerier, mock.Anything, regions).
		Return(nil)

	cardServiceMock := card.NewService(t)
	cardServiceMock.EXPECT().
		CreateCards(mock.Anything, txQuerier).
		Return(nil)

	// Snapshot reader fails
	readerMock := snapshotmock.NewReader(t)
	readerMock.EXPECT().
		GetPublicSnapshot(mock.Anything).
		Return(nil, errors.New("db connection lost"))

	stateStore := handlers.NewStateStore()
	bus := eventbus.NewTestBus()

	svc := creation.NewService(
		poolQuerier,
		bus,
		readerMock,
		stateStore,
		cardServiceMock,
		missionServiceMock,
		playerServiceMock,
		regionServiceMock,
		testStateMetrics(t),
		testGameMetrics(t),
		gamemetrics.NewGameTiming(),
	)

	userCtx := kernelctx.WithUserID(
		kernelctx.WithSpan(t.Context(), tracenoop.Span{}),
		"user-1",
	)

	result, err := svc.CreateGame(userCtx, 0, regions, users)

	// Game creation succeeds despite snapshot read failure
	require.NoError(t, err)
	require.Equal(t, gameID, result)

	// StateStore should be empty (no snapshot to cache)
	require.Nil(t, stateStore.Get(gameID))

	// Lean GameCreated event still emitted
	events := eventbus.EventsOfType[*gameevt.GameCreated](bus)
	require.Len(t, events, 1)
	require.Equal(t, gameID, events[0].GameID())
	require.Nil(t, events[0].PublicSnapshot)
	require.Nil(t, events[0].PrivateSnapshots)
}
