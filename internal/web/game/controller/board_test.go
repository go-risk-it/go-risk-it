package controller_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	ctx2 "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	boardController "github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/game/logic/region"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestBoardControllerImpl_GetBoardState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	regionService := region.NewService(t)

	// Initialize the state under test
	controller := boardController.NewBoardController(regionService)

	// Set up test data
	gameID := int64(1)
	ctx := ctx2.WithGameID(
		ctx2.WithUserID(
			ctx2.WithSpan(t.Context(), noop.Span{}),
			"francesco",
		),
		gameID,
	)

	// Set up expectations for GetRegions method
	regionService.EXPECT().GetRegions(ctx).Return([]sqlc.GetRegionsByGameRow{
		{ExternalReference: "alaska", UserID: "francesco", Troops: 3},
		{ExternalReference: "northwest_territory", UserID: "gabriele", Troops: 3},
		{ExternalReference: "greenland", UserID: "giovanni", Troops: 3},
		{ExternalReference: "alberta", UserID: "francesco", Troops: 3},
	}, nil)

	// Call the method under test
	boardState, err := controller.GetBoardState(ctx)

	// Assert the result
	require.NoError(t, err)
	require.Equal(t, messaging.BoardState{
		Regions: []messaging.Region{
			{ID: "alaska", OwnerID: "francesco", Troops: 3},
			{ID: "northwest_territory", OwnerID: "gabriele", Troops: 3},
			{ID: "greenland", OwnerID: "giovanni", Troops: 3},
			{ID: "alberta", OwnerID: "francesco", Troops: 3},
		},
	}, boardState)
}
