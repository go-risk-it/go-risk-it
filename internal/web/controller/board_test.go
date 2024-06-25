package controller_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/api/game/message"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	boardController "github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/board"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/game/region"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBoardControllerImpl_GetBoardState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	log := zap.NewExample().Sugar()
	boardService := board.NewService(t)
	regionService := region.NewService(t)

	// Initialize the gamestate under test
	controller := boardController.NewBoardController(log, boardService, regionService)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	// Set up expectations for GetRegions method
	regionService.On("GetRegions", ctx, gameID).Return([]sqlc.GetRegionsByGameRow{
		{ExternalReference: "alaska", UserID: "francesco", Troops: 3},
		{ExternalReference: "northwest_territory", UserID: "gabriele", Troops: 3},
		{ExternalReference: "greenland", UserID: "giovanni", Troops: 3},
		{ExternalReference: "alberta", UserID: "francesco", Troops: 3},
	}, nil)

	// Call the method under test
	boardState, err := controller.GetBoardState(ctx, gameID)

	// Assert the result
	require.NoError(t, err)
	require.Equal(t, message.BoardState{
		Regions: []message.Region{
			{ID: "alaska", OwnerID: "francesco", Troops: 3},
			{ID: "northwest_territory", OwnerID: "gabriele", Troops: 3},
			{ID: "greenland", OwnerID: "giovanni", Troops: 3},
			{ID: "alberta", OwnerID: "francesco", Troops: 3},
		},
	}, boardState)

	regionService.AssertExpectations(t)
}
