package game_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/player"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/region"
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
	ctx := context.Background()

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
	mockQuerier.EXPECT().InsertGame(ctx).Return(gameID, nil)
	// mockDB.EXPECT().Begin(qctx).Return()

	playerServiceMock := player.NewService(t)
	playerServiceMock.
		EXPECT().
		CreatePlayers(ctx, mockQuerier, gameID, users).
		Return(players, nil)

	regionServiceMock := region.NewService(t)
	regionServiceMock.
		EXPECT().
		CreateRegions(ctx, mockQuerier, players, regions).
		Return(nil)

	// Initialize the service
	service := game.NewService(
		zap.NewExample().Sugar(),
		mockQuerier,
		playerServiceMock,
		regionServiceMock,
	)

	gameID, err := service.CreateGame(ctx, mockQuerier, gameBoard, users)

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

	// Initialize the service under test
	service := game.NewService(logger, querier, playerService, regionService)

	// Set up test data
	ctx := context.Background()
	gameBoard := &board.Board{} //nolint:exhaustivestruct
	users := []request.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}

	// Set up expectations for InsertGame method
	querier.On("InsertGame", ctx).Return(int64(0), errInsertGame)

	// Call the method under test
	gameID, err := service.CreateGame(ctx, querier, gameBoard, users)

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

	// Initialize the service under test
	service := game.NewService(logger, querier, playerService, regionService)

	// Set up test data
	ctx := context.Background()
	gameBoard := &board.Board{}
	users := []request.Player{
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "user1"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "user2"},
	}

	// Set up expectations for InsertGame method
	querier.On("InsertGame", ctx).Return(int64(1), nil)

	// Set up expectations for CreatePlayers method
	playerService.On("CreatePlayers", ctx, querier, int64(1), users).
		Return(nil, errCreatePlayers)

	// Call the method under test
	gameID, err := service.CreateGame(ctx, querier, gameBoard, users)

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

	// Initialize the service under test
	service := game.NewService(logger, querier, playerService, regionService)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	// Set up expectations for GetGame method
	querier.On("GetGame", ctx, gameID).Return(sqlc.Game{
		ID:    gameID,
		Turn:  3,
		Phase: sqlc.PhaseATTACK,
	}, nil)

	// Call the method under test
	result, err := service.GetGameState(ctx, gameID)

	// Assert the result
	require.NoError(t, err)

	// Verify that the expected methods were called
	require.Equal(t, gameID, result.ID)
}
