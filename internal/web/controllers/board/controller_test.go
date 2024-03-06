package board_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomfran/go-risk-it/internal/api/game/message"
	"github.com/tomfran/go-risk-it/internal/data/sqlc"
	boardController "github.com/tomfran/go-risk-it/internal/web/controllers/board"
	"github.com/tomfran/go-risk-it/mocks/internal_/logic/board"
	"github.com/tomfran/go-risk-it/mocks/internal_/logic/region"
	"go.uber.org/zap"
)

func TestControllerImpl_GetBoardState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	log := zap.NewExample().Sugar()
	boardService := board.NewService(t)
	regionService := region.NewService(t)

	// Initialize the service under test
	controller := boardController.New(log, boardService, regionService)

	// Set up test data
	ctx := context.Background()
	gameID := int64(1)

	// Set up expectations for GetRegions method
	regionService.On("GetRegions", ctx, gameID).Return([]sqlc.GetRegionsByGameRow{
		{ExternalReference: "alaska", PlayerName: "francesco", Troops: 3},
		{ExternalReference: "northwest_territory", PlayerName: "gabriele", Troops: 3},
		{ExternalReference: "greenland", PlayerName: "giovanni", Troops: 3},
		{ExternalReference: "alberta", PlayerName: "francesco", Troops: 3},
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
