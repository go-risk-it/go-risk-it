package game_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	dbmock "github.com/tomfran/go-risk-it/mocks/internal_/db"
	playermock "github.com/tomfran/go-risk-it/mocks/internal_/logic/player"
	regionmock "github.com/tomfran/go-risk-it/mocks/internal_/logic/region"
	"go.uber.org/zap"
)

var (
	errCreatePlayers = errors.New("error inserting players")
	errInsertGame    = errors.New("insert logic error")
)

// creates a logic with a valid board and list of users.
func TestCreateGameWithValidBoardAndUsers(t *testing.T) {
	t.Parallel()

	gameID := int64(1)
	users := []string{"Giovanni", "Gabriele"}
	ctx := context.Background()

	mockQuerier := dbmock.NewQuerier(t)

	players := []db.Player{
		{ID: 420, TurnIndex: 1, GameID: gameID, UserID: "Giovanni"},
		{ID: 69, TurnIndex: 2, GameID: gameID, UserID: "Gabriele"},
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
	mockQuerier.EXPECT().InsertGame(ctx).Return(gameID, nil)

	playerServiceMock := playermock.NewService(t)
	playerServiceMock.
		EXPECT().
		CreatePlayers(ctx, gameID, users).
		Return(players, nil)

	regionServiceMock := regionmock.NewService(t)
	regionServiceMock.
		EXPECT().
		CreateRegions(ctx, mockQuerier, players, regions).
		Return(nil)

	// Initialize the service
	service := game.NewGameService(
		zap.NewExample().Sugar(),
		mockQuerier,
		playerServiceMock,
		regionServiceMock,
	)

	result := service.CreateGame(ctx, gameBoard, users)

	require.NoError(t, result)
}

// returns error if InsertGame method returns an error.
func TestCreateGameInsertGameError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	playerService := playermock.NewService(t)
	regionService := regionmock.NewService(t)
	querier := dbmock.NewQuerier(t)

	// Initialize the service under test
	service := game.NewGameService(logger, querier, playerService, regionService)

	// Set up test data
	ctx := context.Background()
	gameBoard := &board.Board{} //nolint:exhaustivestruct
	users := []string{"user1", "user2"}

	// Set up expectations for InsertGame method
	querier.On("InsertGame", ctx).Return(int64(0), errInsertGame)

	// Call the method under test
	err := service.CreateGame(ctx, gameBoard, users)

	// Assert the result
	require.Error(t, err)
	require.EqualError(t, err, "failed to insert game: insert logic error")

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
}

// returns error if CreatePlayers method returns an error.
func TestCreateGameCreatePlayersError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := dbmock.NewQuerier(t)
	playerService := playermock.NewService(t)
	regionService := regionmock.NewService(t)

	// Initialize the service under test
	service := game.NewGameService(logger, querier, playerService, regionService)

	// Set up test data
	ctx := context.Background()
	gameBoard := &board.Board{}
	users := []string{"user1", "user2"}

	// Set up expectations for InsertGame method
	querier.On("InsertGame", ctx).Return(int64(1), nil)

	// Set up expectations for CreatePlayers method
	playerService.On("CreatePlayers", ctx, int64(1), users).
		Return(nil, errCreatePlayers)

	// Call the method under test
	err := service.CreateGame(ctx, gameBoard, users)

	// Assert the result
	require.Error(t, err)
	require.EqualError(t, err, "failed to create players: error inserting players")

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
	playerService.AssertExpectations(t)
}

func TestGetGameState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := dbmock.NewQuerier(t)
	playerService := playermock.NewService(t)
	regionService := regionmock.NewService(t)

	// Initialize the service under test
	service := game.NewGameService(logger, querier, playerService, regionService)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	// Set up expectations for GetGame method
	querier.On("GetGame", ctx, gameID).Return(db.Game{
		ID:           gameID,
		CurrentTurn:  3,
		CurrentPhase: db.PhaseATTACK,
	}, nil)

	// Call the method under test
	result, err := service.GetGameState(ctx, gameID)

	// Assert the result
	require.NoError(t, err)

	// Verify that the expected methods were called
	require.Equal(t, gameID, result.ID)
}
