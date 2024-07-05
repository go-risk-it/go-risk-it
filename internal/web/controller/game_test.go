package controller_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game/message"
	ctx2 "github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	gameController "github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/board"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/creation"
	gamestate "github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/state"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGameControllerImpl_GetGameState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	log := zap.NewExample().Sugar()
	boardService := board.NewService(t)
	creationService := creation.NewService(t)
	gameService := gamestate.NewService(t)

	// Initialize the state under test
	controller := gameController.NewGameController(boardService, creationService, gameService)

	// Set up test data
	ctx := ctx2.WithGameID(ctx2.WithLog(context.Background(), log), 1)
	gameID := int64(1)

	// Set up expectations for GetGameState method
	gameService.
		EXPECT().
		GetGameState(ctx).
		Return(&state.Game{
			ID:           gameID,
			CurrentTurn:  0,
			CurrentPhase: sqlc.PhaseTypeCARDS,
		}, nil)

	// Call the method under test
	gameState, err := controller.GetGameState(ctx)

	// Assert the result
	require.NoError(t, err)
	require.Equal(t, message.GameState{
		ID:           gameID,
		CurrentTurn:  0,
		CurrentPhase: message.Cards,
	}, gameState)

	gameService.AssertExpectations(t)
}
