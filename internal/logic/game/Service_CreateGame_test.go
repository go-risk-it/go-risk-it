package game_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"github.com/tomfran/go-risk-it/internal/logic/region"
	"go.uber.org/zap"
)

// creates a logic with a valid board and list of users
func TestCreateGameWithValidBoardAndUsers(t *testing.T) {
	t.Parallel()

	gameID := int64(1)
	users := []string{"Giovanni", "Gabriele"}
	ctx := context.Background()

	mockQuerier := db.NewMockQuerier(t)

	players := []db.Player{
		{ID: 420, GameID: gameID, UserID: "Giovanni"},
		{ID: 69, GameID: gameID, UserID: "Gabriele"},
	}

	regions := []board.Region{
		{ExternalReference: 1, Name: "Netherlands", ContinentID: 1},
		{ExternalReference: 2, Name: "Italy", ContinentID: 1},
		{ExternalReference: 3, Name: "Tasin", ContinentID: 2},
		{ExternalReference: 4, Name: "Samon", ContinentID: 3},
	}

	gameBoard := &board.Board{
		Regions:    regions,
		Continents: []board.Continent{},
		Borders:    []board.Border{},
	}

	// setup mocks
	mockQuerier.EXPECT().InsertGame(ctx).Return(1, nil)

	playerServiceMock := player.NewMockService(t)
	playerServiceMock.
		EXPECT().
		CreatePlayers(ctx, mockQuerier, gameID, users).
		Return(players, nil)

	regionServiceMock := region.NewMockService(t)
	regionServiceMock.
		EXPECT().
		CreateRegions(ctx, mockQuerier, players, regions).
		Return(nil)

	// Initialize the service
	service := game.NewGameService(zap.NewExample().Sugar(), playerServiceMock, regionServiceMock)

	result := service.CreateGame(ctx, mockQuerier, gameBoard, users)

	require.NoError(t, result)
}

// returns error if InsertGame method returns an error
func TestCreateGameInsertGameError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	playerService := player.NewMockService(t)
	regionService := region.NewMockService(t)
	querier := db.NewMockQuerier(t)

	// Initialize the service under test
	service := game.NewGameService(logger, playerService, regionService)

	// Set up test data
	ctx := context.Background()
	gameBoard := &board.Board{} //nolint:exhaustivestruct
	users := []string{"user1", "user2"}

	// Set up expectations for InsertGame method
	querier.On("InsertGame", ctx).Return(int64(0), errors.New("insert logic error"))

	// Call the method under test
	err := service.CreateGame(ctx, querier, gameBoard, users)

	// Assert the result
	require.Error(t, err)
	require.EqualError(t, err, "insert logic error")

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
}

// returns error if CreatePlayers method returns an error
func TestCreateGameCreatePlayersError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	playerService := player.NewMockService(t)
	regionService := region.NewMockService(t)

	// Initialize the service under test
	service := game.NewGameService(logger, playerService, regionService)

	// Set up test data
	ctx := context.Background()
	q := db.NewMockQuerier(t)
	gameBoard := &board.Board{}
	users := []string{"user1", "user2"}

	// Set up expectations for InsertGame method
	q.On("InsertGame", ctx).Return(int64(1), nil)

	// Set up expectations for CreatePlayers method
	playerService.On("CreatePlayers", ctx, q, int64(1), users).Return(nil, errors.New("create players error"))

	// Call the method under test
	err := service.CreateGame(ctx, q, gameBoard, users)

	// Assert the result
	require.Error(t, err)
	require.EqualError(t, err, "create players error")

	// Verify that the expected methods were called
	q.AssertExpectations(t)
	playerService.AssertExpectations(t)
}
