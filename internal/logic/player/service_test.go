package player_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	dbmock "github.com/tomfran/go-risk-it/mocks/internal_/db"
	"go.uber.org/zap"
)

func TestServiceImpl_GetPlayers(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := dbmock.NewQuerier(t)

	// Initialize the service under test
	service := player.NewPlayersService(logger, querier)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	player1 := db.Player{
		ID:        1,
		GameID:    gameID,
		UserID:    "user1",
		TurnIndex: 0,
	}
	player2 := db.Player{
		ID:        2,
		GameID:    gameID,
		UserID:    "user2",
		TurnIndex: 1,
	}
	// Set up expectations for GetGame method
	querier.On("GetPlayersByGameId", ctx, gameID).Return([]db.Player{
		player1, player2,
	}, nil)

	// Call the method under test
	result, err := service.GetPlayers(ctx, gameID)

	// Assert the result
	require.NoError(t, err)

	// Verify that the expected methods were called
	require.Len(t, result, 2)
	require.Contains(t, result, player1)
	require.Contains(t, result, player2)
}
