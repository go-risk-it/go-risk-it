package gamestate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/gamestate"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/player"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/stretchr/testify/require"
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
	users := []request.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "Giovanni"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "Gabriele"},
	}
	context := ctx.WithUserID(
		ctx.WithLog(context.Background(), zap.NewExample().Sugar()),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)

	mockQuerier := db.NewQuerier(t)

	players := []sqlc.Player{
		{ID: 420, TurnIndex: 1, GameID: gameID, UserID: "Giovanni"},
		{ID: 69, TurnIndex: 2, GameID: gameID, UserID: "Gabriele"},
	}

	regions := []board.Region{
		{ExternalReference: "netherlands", Name: "Netherlands", Continent: "1"},
		{ExternalReference: "italy", Name: "Italy", Continent: "1"},
		{ExternalReference: "tasin", Name: "Tasin", Continent: "2"},
		{ExternalReference: "samon", Name: "Samon", Continent: "3"},
	}

	gameBoard := &board.Board{
		Regions:    regions,
		Continents: []board.Continent{},
		Borders:    []board.Border{},
	}

	// setup mocks
	mockQuerier.EXPECT().InsertGame(context, int64(3)).Return(sqlc.Game{
		ID:               gameID,
		Turn:             1,
		Phase:            sqlc.PhaseDEPLOY,
		DeployableTroops: 3,
	}, nil)
	// mockDB.EXPECT().Begin(qctx).Return()

	playerServiceMock := player.NewService(t)
	playerServiceMock.
		EXPECT().
		CreatePlayers(context, mockQuerier, gameID, users).
		Return(players, nil)

	regionServiceMock := region.NewService(t)
	regionServiceMock.
		EXPECT().
		CreateRegions(context, mockQuerier, players, regions).
		Return(nil)

	// Initialize the gamestate
	service := gamestate.NewService(
		mockQuerier,
		playerServiceMock,
		regionServiceMock,
	)

	gameID, err := service.CreateGameQ(context, mockQuerier, gameBoard, users)

	require.NoError(t, err)
	require.Equal(t, int64(1), gameID)
}

// returns error if InsertGame method returns an error.
func TestServiceImpl_CreateGame_InsertGameError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	playerService := player.NewService(t)
	regionService := region.NewService(t)
	querier := db.NewQuerier(t)

	// Initialize the gamestate under test
	service := gamestate.NewService(querier, playerService, regionService)

	// Set up test data
	ctx := ctx.WithUserID(
		ctx.WithLog(context.Background(), logger),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)
	gameBoard := &board.Board{} //nolint:exhaustivestruct
	users := []request.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}

	// Set up expectations for InsertGame method
	querier.On("InsertGame", ctx, int64(3)).Return(sqlc.Game{}, errInsertGame)

	// Call the method under test
	gameID, err := service.CreateGameQ(ctx, querier, gameBoard, users)

	// Assert the result
	require.Error(t, err)
	require.EqualError(t, err, "failed to insert game: insert logic error")
	require.Equal(t, int64(-1), gameID)

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
}

// returns error if CreatePlayers method returns an error.
func TestServiceImpl_CreateGame_CreatePlayersError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := db.NewQuerier(t)
	playerService := player.NewService(t)
	regionService := region.NewService(t)

	// Initialize the gamestate under test
	service := gamestate.NewService(querier, playerService, regionService)

	// Set up test data
	ctx := ctx.WithUserID(
		ctx.WithLog(context.Background(), logger),
		"dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
	)
	gameBoard := &board.Board{}
	users := []request.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}

	// Set up expectations for InsertGame method
	querier.On("InsertGame", ctx, int64(3)).Return(sqlc.Game{
		ID: 1,
	}, nil)

	// Set up expectations for CreatePlayers method
	playerService.On("CreatePlayers", ctx, querier, int64(1), users).
		Return(nil, errCreatePlayers)

	// Call the method under test
	gameID, err := service.CreateGameQ(ctx, querier, gameBoard, users)

	// Assert the result
	require.Error(t, err)
	require.EqualError(t, err, "failed to create players: error inserting players")
	require.Equal(t, int64(-1), gameID)

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
	playerService.AssertExpectations(t)
}

func TestServiceImpl_GetGameState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := db.NewQuerier(t)
	playerService := player.NewService(t)
	regionService := region.NewService(t)

	// Initialize the gamestate under test
	service := gamestate.NewService(querier, playerService, regionService)

	// Set up test data
	gameID := int64(1)
	ctx := ctx.WithGameID(ctx.WithLog(context.Background(), logger), gameID)

	// Set up expectations for GetGame method
	querier.On("GetGame", ctx, gameID).Return(sqlc.Game{
		ID:    gameID,
		Turn:  3,
		Phase: sqlc.PhaseATTACK,
	}, nil)

	// Call the method under test
	result, err := service.GetGameState(ctx)

	// Assert the result
	require.NoError(t, err)

	// Verify that the expected methods were called
	require.Equal(t, gameID, result.ID)
}
