package player_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	sqlc "github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"github.com/tomfran/go-risk-it/mocks/internal_/data/db"
	"go.uber.org/zap"
)

func TestServiceImpl_GetPlayers(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := db.NewQuerier(t)

	// Initialize the service under test
	service := player.NewPlayersService(logger, querier)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	player1 := sqlc.Player{
		ID:        1,
		GameID:    gameID,
		UserID:    "user1",
		TurnIndex: 0,
	}
	player2 := sqlc.Player{
		ID:        2,
		GameID:    gameID,
		UserID:    "user2",
		TurnIndex: 1,
	}
	// Set up expectations for GetGame method
	querier.On("GetPlayersByGame", ctx, gameID).Return([]sqlc.Player{
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
