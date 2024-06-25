package controller_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game/message"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	gameController "github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/board"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/gamestate"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/player"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGameControllerImpl_GetGameState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	log := zap.NewExample().Sugar()
	gameService := gamestate.NewService(t)
	boardService := board.NewService(t)
	playerService := player.NewService(t)

	// Initialize the gamestate under test
	controller := gameController.NewGameController(log, gameService, boardService, playerService)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	// Set up expectations for GetGameState method
	gameService.On("GetGameState", ctx, gameID).Return(&sqlc.Game{
		ID:    gameID,
		Turn:  0,
		Phase: "CARDS",
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
