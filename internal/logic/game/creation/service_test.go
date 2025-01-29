package creation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/creation"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/card"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/player"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

var (
	errCreatePlayers = errors.New("error inserting players")
	errInsertGame    = errors.New("insert logic error")
)

// creates a game with a valid board and list of users.
func TestServiceImpl_CreateGame_WithValidBoardAndUsers(t *testing.T) {
	t.Parallel()

	gameID := int64(1)
	phaseID := int64(1)
	users := []request.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "Giovanni"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "Gabriele"},
	}
	context := ctx.WithUserID(
		ctx.WithSpan(ctx.WithLog(context.Background(), zap.NewExample().Sugar()), noop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)

	mockQuerier := db.NewQuerier(t)

	players := []sqlc.Player{
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
	mockQuerier.EXPECT().InsertGame(context).Return(sqlc.Game{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: 1, Valid: true},
	}, nil)

	gameContext := ctx.WithGameID(context, gameID)

	mockQuerier.EXPECT().InsertPhase(gameContext, sqlc.InsertPhaseParams{
		GameID: gameID,
		Type:   sqlc.PhaseTypeDEPLOY,
		Turn:   0,
	}).Return(sqlc.Phase{ID: phaseID}, nil)

	mockQuerier.EXPECT().SetGamePhase(gameContext, sqlc.SetGamePhaseParams{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: phaseID, Valid: true},
	}).Return(nil)

	mockQuerier.EXPECT().InsertDeployPhase(gameContext, sqlc.InsertDeployPhaseParams{
		PhaseID:          phaseID,
		DeployableTroops: int64(3),
	}).Return(sqlc.DeployPhase{ID: 1}, nil)

	playerServiceMock := player.NewService(t)
	playerServiceMock.
		EXPECT().
		CreatePlayersQ(gameContext, mockQuerier, gameID, users).
		Return(players, nil)

	missionServiceMock := mission.NewService(t)
	missionServiceMock.
		EXPECT().
		CreateMissionsQ(gameContext, mockQuerier, players).
		Return(nil)

	regionServiceMock := region.NewService(t)
	regionServiceMock.
		EXPECT().
		CreateRegionsQ(gameContext, mockQuerier, players, regions).
		Return(nil)

	cardServiceMock := card.NewService(t)
	cardServiceMock.
		EXPECT().
		CreateCardsQ(gameContext, mockQuerier).
		Return(nil)

	// Initialize the state
	service := creation.NewService(
		mockQuerier,
		cardServiceMock,
		missionServiceMock,
		playerServiceMock,
		regionServiceMock,
	)

	gameID, err := service.CreateGameQ(context, mockQuerier, regions, users)

	require.NoError(t, err)
	require.Equal(t, int64(1), gameID)
}

// returns error if InsertGame method returns an error.
func TestServiceImpl_CreateGame_InsertGameError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	cardService := card.NewService(t)
	missionService := mission.NewService(t)
	playerService := player.NewService(t)
	regionService := region.NewService(t)
	querier := db.NewQuerier(t)

	// Initialize the state under test
	service := creation.NewService(
		querier,
		cardService,
		missionService,
		playerService,
		regionService,
	)

	// Set up test data
	ctx := ctx.WithUserID(
		ctx.WithSpan(ctx.WithLog(context.Background(), zap.NewExample().Sugar()), noop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)
	users := []request.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}

	// Set up expectations for InsertGame method
	querier.
		EXPECT().
		InsertGame(ctx).Return(sqlc.Game{}, errInsertGame)

	// Call the method under test
	gameID, err := service.CreateGameQ(ctx, querier, []string{}, users)

	// Assert the result
	require.Error(t, err)
	require.EqualError(t, err, "failed to insert game: insert logic error")
	require.Equal(t, int64(-1), gameID)

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
}

// returns error if CreatePlayersQ method returns an error.
func TestServiceImpl_CreateGame_CreatePlayersError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	querier := db.NewQuerier(t)
	cardService := card.NewService(t)
	missionService := mission.NewService(t)
	playerService := player.NewService(t)
	regionService := region.NewService(t)

	// Initialize the state under test
	service := creation.NewService(
		querier,
		cardService,
		missionService,
		playerService,
		regionService,
	)

	// Set up test data
	context := ctx.WithUserID(
		ctx.WithSpan(ctx.WithLog(context.Background(), zap.NewExample().Sugar()), noop.Span{}),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)
	users := []request.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}
	gameID := int64(1)

	// Set up expectations for InsertGame method
	querier.
		EXPECT().
		InsertGame(context).
		Return(sqlc.Game{
			ID: gameID,
		}, nil)

	gameContext := ctx.WithGameID(context, gameID)

	// Set up expectations for CreatePlayersQ method
	playerService.
		EXPECT().
		CreatePlayersQ(gameContext, querier, int64(1), users).
		Return(nil, errCreatePlayers)

	// Call the method under test
	gameID, err := service.CreateGameQ(context, querier, []string{}, users)

	// Assert the result
	require.Error(t, err)
	require.EqualError(t, err, "failed to create players: error inserting players")
	require.Equal(t, int64(-1), gameID)

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
	playerService.AssertExpectations(t)
}
