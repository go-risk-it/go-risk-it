package creation_test

import (
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/creation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/timing"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/card"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/mission"
	playermock "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/player"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

var (
	errCreatePlayers = errors.New("error inserting players")
	errInsertGame    = errors.New("insert logic error")
)

func testMetrics(t *testing.T) *metrics.Metrics {
	t.Helper()

	m, err := metrics.NewMetrics(noop.Meter{})
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
	context := ctx.WithUserID(
		ctx.WithSpan(t.Context(), tracenoop.Span{}),
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
	mockQuerier.EXPECT().InsertGame(context).Return(sqlc.GameGame{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: 1, Valid: true},
	}, nil)

	gameContext := ctx.WithGameID(context, gameID)

	mockQuerier.EXPECT().InsertPhase(gameContext, sqlc.InsertPhaseParams{
		GameID: gameID,
		Type:   sqlc.GamePhaseTypeDEPLOY,
		Turn:   0,
	}).Return(sqlc.GamePhase{ID: phaseID}, nil)

	mockQuerier.EXPECT().SetGamePhase(gameContext, sqlc.SetGamePhaseParams{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: phaseID, Valid: true},
	}).Return(nil)

	mockQuerier.EXPECT().InsertDeployPhase(gameContext, sqlc.InsertDeployPhaseParams{
		PhaseID:          phaseID,
		DeployableTroops: int64(3),
	}).Return(sqlc.GameDeployPhase{ID: 1}, nil)

	playerServiceMock := playermock.NewService(t)
	playerServiceMock.
		EXPECT().
		CreatePlayers(gameContext, mockQuerier, gameID, users).
		Return(players, nil)

	missionServiceMock := mission.NewService(t)
	missionServiceMock.
		EXPECT().
		CreateMissions(gameContext, mockQuerier, players).
		Return(nil)

	regionServiceMock := region.NewService(t)
	regionServiceMock.
		EXPECT().
		CreateRegions(gameContext, mockQuerier, players, regions).
		Return(nil)

	cardServiceMock := card.NewService(t)
	cardServiceMock.
		EXPECT().
		CreateCards(gameContext, mockQuerier).
		Return(nil)

	// Initialize the state
	service := creation.NewService(
		mockQuerier,
		eventbus.NewTestBus(),
		cardServiceMock,
		missionServiceMock,
		playerServiceMock,
		regionServiceMock,
		testMetrics(t),
		timing.NewGameTiming(),
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
		cardService,
		missionService,
		playerService,
		regionService,
		testMetrics(t),
		timing.NewGameTiming(),
	)

	// Set up test data
	ctx := ctx.WithUserID(
		ctx.WithSpan(t.Context(), tracenoop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)
	users := []player.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}

	// Set up expectations for InsertGame method
	querier.
		EXPECT().
		InsertGame(ctx).Return(sqlc.GameGame{}, errInsertGame)

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
		cardService,
		missionService,
		playerService,
		regionService,
		testMetrics(t),
		timing.NewGameTiming(),
	)

	// Set up test data
	context := ctx.WithUserID(
		ctx.WithSpan(t.Context(), tracenoop.Span{}),
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
		InsertGame(context).
		Return(sqlc.GameGame{
			ID: gameID,
		}, nil)

	gameContext := ctx.WithGameID(context, gameID)

	// Set up expectations for CreatePlayers method
	playerService.
		EXPECT().
		CreatePlayers(gameContext, querier, int64(1), users).
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
	context := ctx.WithUserID(
		ctx.WithSpan(t.Context(), tracenoop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)

	mockQuerier := db.NewQuerier(t)

	players := []sqlc.GamePlayer{
		{ID: 420, TurnIndex: 1, GameID: gameID, UserID: "Giovanni"},
		{ID: 69, TurnIndex: 2, GameID: gameID, UserID: "Gabriele"},
	}

	regions := []string{"netherlands", "italy", "tasin", "samon"}

	mockQuerier.EXPECT().InsertGame(context).Return(sqlc.GameGame{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: 1, Valid: true},
	}, nil)

	gameContext := ctx.WithGameID(context, gameID)

	mockQuerier.EXPECT().InsertPhase(gameContext, sqlc.InsertPhaseParams{
		GameID: gameID,
		Type:   sqlc.GamePhaseTypeDEPLOY,
		Turn:   0,
	}).Return(sqlc.GamePhase{ID: phaseID}, nil)

	mockQuerier.EXPECT().SetGamePhase(gameContext, sqlc.SetGamePhaseParams{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: phaseID, Valid: true},
	}).Return(nil)

	mockQuerier.EXPECT().InsertDeployPhase(gameContext, sqlc.InsertDeployPhaseParams{
		PhaseID:          phaseID,
		DeployableTroops: int64(3),
	}).Return(sqlc.GameDeployPhase{ID: 1}, nil)

	playerServiceMock := playermock.NewService(t)
	playerServiceMock.EXPECT().
		CreatePlayers(gameContext, mockQuerier, gameID, users).
		Return(players, nil)

	missionServiceMock := mission.NewService(t)
	missionServiceMock.EXPECT().
		CreateMissions(gameContext, mockQuerier, players).
		Return(nil)

	regionServiceMock := region.NewService(t)
	regionServiceMock.EXPECT().
		CreateRegions(gameContext, mockQuerier, players, regions).
		Return(nil)

	cardServiceMock := card.NewService(t)
	cardServiceMock.EXPECT().
		CreateCards(gameContext, mockQuerier).
		Return(nil)

	bus := eventbus.NewTestBus()
	service := creation.NewService(
		mockQuerier,
		bus,
		cardServiceMock,
		missionServiceMock,
		playerServiceMock,
		regionServiceMock,
		testMetrics(t),
		timing.NewGameTiming(),
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
		card.NewService(t),
		mission.NewService(t),
		playermock.NewService(t),
		region.NewService(t),
		testMetrics(t),
		timing.NewGameTiming(),
	)

	userCtx := ctx.WithUserID(
		ctx.WithSpan(t.Context(), tracenoop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)
	users := []player.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}

	querier.EXPECT().InsertGame(userCtx).Return(sqlc.GameGame{}, errInsertGame)

	_, err := service.CreateGameWithQuerier(userCtx, querier, []string{}, users)

	require.Error(t, err)
	require.Empty(t, bus.Events(), "no events should be emitted when CreateGameWithQuerier fails")
}
