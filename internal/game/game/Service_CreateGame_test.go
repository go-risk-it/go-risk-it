package game

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/game/board"
	"github.com/tomfran/go-risk-it/internal/game/player"
	"github.com/tomfran/go-risk-it/internal/game/region"
	"go.uber.org/zap"
	"testing"
)

// creates a game with a valid board and list of users
func TestCreateGameWithValidBoardAndUsers(t *testing.T) {
	var gameId int64 = 1
	users := []string{"Giovanni", "Gabriele"}
	ctx := context.Background()
	mockQuerier := db.NewMockQuerier(t)

	players := []db.Player{
		{420, gameId, "Giovanni"},
		{69, gameId, "Gabriele"},
	}

	regions := []board.Region{
		{1, "Netherlands", 1},
		{2, "Italy", 1},
		{3, "Tasin", 2},
		{4, "Samon", 3},
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
		CreatePlayers(ctx, mockQuerier, gameId, users).
		Return(players, nil)

	regionServiceMock := region.NewMockService(t)
	regionServiceMock.
		EXPECT().
		CreateRegions(ctx, mockQuerier, players, regions).
		Return(nil)

	// Initialize the service
	service := &ServiceImpl{
		log:           zap.NewExample().Sugar(),
		playerService: playerServiceMock,
		regionService: regionServiceMock,
	}

	result := service.CreateGame(ctx, mockQuerier, gameBoard, users)

	assert.NoError(t, result)
}

// returns error if InsertGame method returns an error
func TestCreateGameInsertGameError(t *testing.T) {
	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	playerService := player.NewMockService(t)
	regionService := region.NewMockService(t)
	querier := db.NewMockQuerier(t)

	// Initialize the service under test
	service := &ServiceImpl{
		log:           logger,
		playerService: playerService,
		regionService: regionService,
	}

	// Set up test data
	ctx := context.Background()
	board := &board.Board{}
	users := []string{"user1", "user2"}

	// Set up expectations for InsertGame method
	querier.On("InsertGame", ctx).Return(int64(0), errors.New("insert game error"))

	// Call the method under test
	err := service.CreateGame(ctx, querier, board, users)

	// Assert the result
	assert.Error(t, err)
	assert.EqualError(t, err, "insert game error")

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
}

// returns error if CreatePlayers method returns an error
func TestCreateGameCreatePlayersError(t *testing.T) {
	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	playerService := player.NewMockService(t)
	regionService := region.NewMockService(t)

	// Initialize the service under test
	service := &ServiceImpl{
		log:           logger,
		playerService: playerService,
		regionService: regionService,
	}

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
	assert.Error(t, err)
	assert.EqualError(t, err, "create players error")

	// Verify that the expected methods were called
	q.AssertExpectations(t)
	playerService.AssertExpectations(t)
}
