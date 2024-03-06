package game_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomfran/go-risk-it/internal/api/game/message"
	"github.com/tomfran/go-risk-it/internal/data/sqlc"
	gameController "github.com/tomfran/go-risk-it/internal/web/controllers/game"
	"github.com/tomfran/go-risk-it/mocks/internal_/logic/board"
	"github.com/tomfran/go-risk-it/mocks/internal_/logic/game"
	"github.com/tomfran/go-risk-it/mocks/internal_/logic/player"
	"go.uber.org/zap"
)

func TestControllerImpl_GetGameState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	log := zap.NewExample().Sugar()
	gameService := game.NewService(t)
	boardService := board.NewService(t)
	playerService := player.NewService(t)

	// Initialize the service under test
	controller := gameController.New(log, gameService, boardService, playerService)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	// Set up expectations for GetGameState method
	gameService.On("GetGameState", ctx, gameID).Return(&sqlc.Game{
		ID:           gameID,
		CurrentTurn:  0,
		CurrentPhase: "CARDS",
	}, nil)

	// Call the method under test
	gameState, err := controller.GetGameState(ctx, gameID)

	// Assert the result
	require.NoError(t, err)
	require.Equal(t, message.GameState{
		GameID:       gameID,
		CurrentTurn:  0,
		CurrentPhase: "CARDS",
	}, gameState)

	gameService.AssertExpectations(t)
}
