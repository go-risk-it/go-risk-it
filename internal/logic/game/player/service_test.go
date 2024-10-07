package player_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	ctx2 "github.com/go-risk-it/go-risk-it/internal/ctx"
	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	errInsertPlayers    = errors.New("error inserting players")
	errGetPlayersByGame = errors.New("error getting players")
)

func TestServiceImpl_CreatePlayers_WithValidData(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := db.NewQuerier(t)

	// Initialize the state under test
	service := player.NewService(querier)

	// Set up test data
	gameID := int64(1)
	ctx := ctx2.WithGameID(ctx2.WithUserID(
		ctx2.WithLog(context.Background(), logger),
		"5a4fde41-4a68-4625-b42b-a9f5f938b394",
	), gameID)
	users := []request.Player{
		{UserID: "5a4fde41-4a68-4625-b42b-a9f5f938b394", Name: "francesco"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "gabriele"},
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "giovanni"},
	}

	// Set up expectations for InsertPlayers method
	querier.On("InsertPlayers", ctx, []sqlc.InsertPlayersParams{
		{
			GameID:    gameID,
			UserID:    "5a4fde41-4a68-4625-b42b-a9f5f938b394",
			Name:      "francesco",
			TurnIndex: 0,
		},
		{
			GameID:    gameID,
			UserID:    "dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
			Name:      "gabriele",
			TurnIndex: 1,
		},
		{
			GameID:    gameID,
			UserID:    "fc497971-de4d-49c2-842a-4af62ec9e858",
			Name:      "giovanni",
			TurnIndex: 2,
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

	// Initialize the state under test
	service := player.NewService(querier)

	// Set up test data
	gameID := int64(1)
	ctx := ctx2.WithGameID(ctx2.WithUserID(
		ctx2.WithLog(context.Background(), logger),
		"5a4fde41-4a68-4625-b42b-a9f5f938b394",
	), gameID)
	users := []request.Player{
		{UserID: "5a4fde41-4a68-4625-b42b-a9f5f938b394", Name: "francesco"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "gabriele"},
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "giovanni"},
	}

	// Set up expectations for InsertPlayers method
	querier.On("InsertPlayers", ctx, []sqlc.InsertPlayersParams{
		{
			GameID:    gameID,
			UserID:    "5a4fde41-4a68-4625-b42b-a9f5f938b394",
			Name:      "francesco",
			TurnIndex: 0,
		},
		{
			GameID:    gameID,
			UserID:    "dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
			Name:      "gabriele",
			TurnIndex: 1,
		},
		{
			GameID:    gameID,
			UserID:    "fc497971-de4d-49c2-842a-4af62ec9e858",
			Name:      "giovanni",
			TurnIndex: 2,
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

	// Initialize the state under test
	service := player.NewService(querier)

	// Set up test data
	gameID := int64(1)
	ctx := ctx2.WithGameID(ctx2.WithUserID(
		ctx2.WithLog(context.Background(), logger),
		"5a4fde41-4a68-4625-b42b-a9f5f938b394",
	), gameID)
	users := []request.Player{
		{UserID: "5a4fde41-4a68-4625-b42b-a9f5f938b394", Name: "francesco"},
		{UserID: "dc2dabc6-ca5b-41af-8cb4-8eb768f13258", Name: "gabriele"},
		{UserID: "fc497971-de4d-49c2-842a-4af62ec9e858", Name: "giovanni"},
	}

	// Set up expectations for InsertPlayers method
	querier.On("InsertPlayers", ctx, []sqlc.InsertPlayersParams{
		{
			GameID:    gameID,
			UserID:    "5a4fde41-4a68-4625-b42b-a9f5f938b394",
			Name:      "francesco",
			TurnIndex: 0,
		},
		{
			GameID:    gameID,
			UserID:    "dc2dabc6-ca5b-41af-8cb4-8eb768f13258",
			Name:      "gabriele",
			TurnIndex: 1,
		},
		{
			GameID:    gameID,
			UserID:    "fc497971-de4d-49c2-842a-4af62ec9e858",
			Name:      "giovanni",
			TurnIndex: 2,
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
