package controller_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	ctx2 "github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	playerController "github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/player"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestControllerImpl_GetPlayerState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	playerService := player.NewService(t)

	// Initialize the state under test
	controller := playerController.NewPlayerController(playerService)

	// Set up test data
	gameID := int64(1)
	ctx := ctx2.WithGameID(
		ctx2.WithUserID(ctx2.WithLog(context.Background(), logger), "francesco"),
		gameID,
	)

	// Set up expectations for GetPlayersState method
	playerService.EXPECT().GetPlayersState(ctx).Return([]sqlc.GetPlayersStateRow{
		{Name: "francesco", UserID: "user1", TurnIndex: 0, CardCount: 0},
		{Name: "gabriele", UserID: "user2", TurnIndex: 1, CardCount: 0},
	}, nil)

	// Call the method under test
	playerState, err := controller.GetPlayerState(ctx)

	// Assert the result
	require.NoError(t, err)
	require.Equal(t, messaging.PlayersState{
		Players: []messaging.Player{
			{UserID: "user1", Name: "francesco", Index: 0, CardCount: 0},
			{UserID: "user2", Name: "gabriele", Index: 1, CardCount: 0},
		},
	}, playerState)

	playerService.AssertExpectations(t)
}
