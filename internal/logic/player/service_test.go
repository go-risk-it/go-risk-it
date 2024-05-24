package player_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
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
	users := []request.Player{
		{UserID: "5a4fde41-4a68-4625-b42b-a9f5f938b394", Name: "francesco"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "gabriele"},
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "giovanni"},
	}

	// Set up expectations for InsertPlayers method
	querier.On("InsertPlayers", ctx, []sqlc.InsertPlayersParams{
		{
			GameID:           gameID,
			UserID:           "5a4fde41-4a68-4625-b42b-a9f5f938b394",
			Name:             "francesco",
			TurnIndex:        0,
			DeployableTroops: 5,
		},
		{
			GameID:           gameID,
			UserID:           "dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
			Name:             "gabriele",
			TurnIndex:        1,
			DeployableTroops: 5,
		},
		{
			GameID:           gameID,
			UserID:           "fc497971-de4d-49c2-842a-4af62ec9e858",
			Name:             "giovanni",
			TurnIndex:        2,
			DeployableTroops: 5,
		},
	}).Return(int64(2), nil)

	querier.On("GetPlayersByGame", ctx, gameID).Return([]sqlc.Player{
		{
			ID:        1,
			GameID:    gameID,
			UserID:    "5a4fde41-4a68-4625-b42b-a9f5f938b394",
			Name:      "francesco",
			TurnIndex: 0,
		},
		{
			ID:        2,
			GameID:    gameID,
			UserID:    "dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
			Name:      "gabriele",
			TurnIndex: 1,
		},
		{
			ID:        3,
			GameID:    gameID,
			UserID:    "fc497971-de4d-49c2-842a-4af62ec9e858",
			Name:      "giovanni",
			TurnIndex: 2,
		},
	}, nil)

	// Call the method under test
	players, err := service.CreatePlayers(ctx, querier, gameID, users)

	// Assert the result
	require.NoError(t, err)
	require.Len(t, players, 3)
	require.Equal(t, "5a4fde41-4a68-4625-b42b-a9f5f938b394", players[0].UserID)
	require.Equal(t, "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", players[1].UserID)
	require.Equal(t, "fc497971-de4d-49c2-842a-4af62ec9e858", players[2].UserID)

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
	users := []request.Player{
		{UserID: "5a4fde41-4a68-4625-b42b-a9f5f938b394", Name: "francesco"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "gabriele"},
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "giovanni"},
	}

	// Set up expectations for InsertPlayers method
	querier.On("InsertPlayers", ctx, []sqlc.InsertPlayersParams{
		{
			GameID:           gameID,
			UserID:           "5a4fde41-4a68-4625-b42b-a9f5f938b394",
			Name:             "francesco",
			TurnIndex:        0,
			DeployableTroops: 5,
		},
		{
			GameID:           gameID,
			UserID:           "dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
			Name:             "gabriele",
			TurnIndex:        1,
			DeployableTroops: 5,
		},
		{
			GameID:           gameID,
			UserID:           "fc497971-de4d-49c2-842a-4af62ec9e858",
			Name:             "giovanni",
			TurnIndex:        2,
			DeployableTroops: 5,
		},
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
	users := []request.Player{
		{UserID: "5a4fde41-4a68-4625-b42b-a9f5f938b394", Name: "francesco"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "gabriele"},
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "giovanni"},
	}

	// Set up expectations for InsertPlayers method
	querier.On("InsertPlayers", ctx, []sqlc.InsertPlayersParams{
		{
			GameID:           gameID,
			UserID:           "5a4fde41-4a68-4625-b42b-a9f5f938b394",
			Name:             "francesco",
			TurnIndex:        0,
			DeployableTroops: 5,
		},
		{
			GameID:           gameID,
			UserID:           "dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
			Name:             "gabriele",
			TurnIndex:        1,
			DeployableTroops: 5,
		},
		{
			GameID:           gameID,
			UserID:           "fc497971-de4d-49c2-842a-4af62ec9e858",
			Name:             "giovanni",
			TurnIndex:        2,
			DeployableTroops: 5,
		},
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
