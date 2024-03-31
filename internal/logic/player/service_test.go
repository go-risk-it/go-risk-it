package player_test

import (
	"context"
	"errors"
	"testing"

	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	errInsertPlayers    = errors.New("error inserting players")
	errGetPlayersByGame = errors.New("error getting players")
)

func TestServiceImpl_GetPlayersByGame(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := db.NewQuerier(t)

	// Initialize the service under test
	service := player.NewService(logger, querier)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	player1 := sqlc.Player{
		ID:        1,
		GameID:    gameID,
		UserID:    "francesco",
		TurnIndex: 0,
	}
	player2 := sqlc.Player{
		ID:        2,
		GameID:    gameID,
		UserID:    "gabriele",
		TurnIndex: 1,
	}
	player3 := sqlc.Player{
		ID:        3,
		GameID:    gameID,
		UserID:    "giovanni",
		TurnIndex: 2,
	}
	// Set up expectations for GetGame method
	querier.On("GetPlayersByGame", ctx, gameID).Return([]sqlc.Player{
		player1, player2, player3,
	}, nil)

	// Call the method under test
	result, err := service.GetPlayers(ctx, gameID)

	// Assert the result
	require.NoError(t, err)

	// Verify that the expected methods were called
	require.Len(t, result, 3)
	require.Contains(t, result, player1)
	require.Contains(t, result, player2)
	require.Contains(t, result, player3)
}

func TestServiceImpl_GetPlayersByGame_WithError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := db.NewQuerier(t)

	// Initialize the service under test
	service := player.NewService(logger, querier)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	// Set up expectations for GetGame method
	querier.On("GetPlayersByGame", ctx, gameID).Return(nil, errGetPlayersByGame)

	// Call the method under test
	result, err := service.GetPlayers(ctx, gameID)

	// Assert the result
	require.Error(t, err)
	require.Nil(t, result)
}

func TestServiceImpl_CreatePlayers_WithValidData(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := db.NewQuerier(t)

	// Initialize the service under test
	service := player.NewService(logger, querier)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)
	users := []string{"francesco", "gabriele", "giovanni"}

	// Set up expectations for InsertPlayers method
	querier.On("InsertPlayers", ctx, []sqlc.InsertPlayersParams{
		{GameID: gameID, UserID: "francesco", TurnIndex: 0},
		{GameID: gameID, UserID: "gabriele", TurnIndex: 1},
		{GameID: gameID, UserID: "giovanni", TurnIndex: 2},
	}).Return(int64(2), nil)

	querier.On("GetPlayersByGame", ctx, gameID).Return([]sqlc.Player{
		{ID: 1, GameID: gameID, UserID: "francesco", TurnIndex: 0},
		{ID: 2, GameID: gameID, UserID: "gabriele", TurnIndex: 1},
		{ID: 3, GameID: gameID, UserID: "giovanni", TurnIndex: 2},
	}, nil)

	// Call the method under test
	players, err := service.CreatePlayers(ctx, querier, gameID, users)

	// Assert the result
	require.NoError(t, err)
	require.Len(t, players, 3)
	require.Equal(t, "francesco", players[0].UserID)
	require.Equal(t, "gabriele", players[1].UserID)
	require.Equal(t, "giovanni", players[2].UserID)

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
}

func TestServiceImpl_CreatePlayers_InsertPlayersError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := db.NewQuerier(t)

	// Initialize the service under test
	service := player.NewService(logger, querier)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)
	users := []string{"francesco", "gabriele", "giovanni"}

	// Set up expectations for InsertPlayers method
	querier.On("InsertPlayers", ctx, []sqlc.InsertPlayersParams{
		{GameID: gameID, UserID: "francesco", TurnIndex: 0},
		{GameID: gameID, UserID: "gabriele", TurnIndex: 1},
		{GameID: gameID, UserID: "giovanni", TurnIndex: 2},
	}).Return(int64(0), errInsertPlayers)

	// Call the method under test
	players, err := service.CreatePlayers(ctx, querier, gameID, users)

	// Assert the result
	require.Error(t, err)
	require.Nil(t, players)

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
}

func TestServiceImpl_CreatePlayers_GetPlayersByGameError(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := db.NewQuerier(t)

	// Initialize the service under test
	service := player.NewService(logger, querier)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)
	users := []string{"francesco", "gabriele", "giovanni"}

	// Set up expectations for InsertPlayers method
	querier.On("InsertPlayers", ctx, []sqlc.InsertPlayersParams{
		{GameID: gameID, UserID: "francesco", TurnIndex: 0},
		{GameID: gameID, UserID: "gabriele", TurnIndex: 1},
		{GameID: gameID, UserID: "giovanni", TurnIndex: 2},
	}).Return(int64(2), nil)

	querier.On("GetPlayersByGame", ctx, gameID).Return(nil, errGetPlayersByGame)

	// Call the method under test
	players, err := service.CreatePlayers(ctx, querier, gameID, users)

	// Assert the result
	require.Error(t, err)
	require.Nil(t, players)

	// Verify that the expected methods were called
	querier.AssertExpectations(t)
}
