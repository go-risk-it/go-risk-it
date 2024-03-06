package player_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomfran/go-risk-it/internal/api/game/message"
	"github.com/tomfran/go-risk-it/internal/data/sqlc"
	playerController "github.com/tomfran/go-risk-it/internal/web/controllers/player"
	"github.com/tomfran/go-risk-it/mocks/internal_/logic/player"
	"go.uber.org/zap"
)

func TestControllerImpl_GetPlayerState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	playerService := player.NewService(t)

	// Initialize the service under test
	controller := playerController.New(logger, playerService)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	// Set up expectations for GetPlayers method
	playerService.On("GetPlayers", ctx, gameID).Return([]sqlc.Player{
		{ID: 1, GameID: gameID, UserID: "user1", TurnIndex: 0},
		{ID: 2, GameID: gameID, UserID: "user2", TurnIndex: 1},
	}, nil)

	// Call the method under test
	playerState, err := controller.GetPlayerState(ctx, gameID)

	// Assert the result
	require.NoError(t, err)
	require.Equal(t, message.PlayersState{
		Players: []message.Player{
			{ID: "user1", Index: 0},
			{ID: "user2", Index: 1},
		},
	}, playerState)

	playerService.AssertExpectations(t)
}
