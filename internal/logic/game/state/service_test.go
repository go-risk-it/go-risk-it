package state_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/db"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServiceImpl_GetGameState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	logger := zap.NewExample().Sugar()
	querier := db.NewQuerier(t)

	// Initialize the state under test
	service := state.NewService(querier)

	// Set up test data
	gameID := int64(1)
	ctx := ctx.WithGameID(ctx.WithLog(context.Background(), logger), gameID)

	// Set up expectations for GetGame method
	querier.EXPECT().GetGame(ctx, gameID).Return(sqlc.GetGameRow{
		ID:           gameID,
		CurrentPhase: sqlc.PhaseTypeATTACK,
		Turn:         3,
	}, nil)

	// Call the method under test
	result, err := service.GetGameState(ctx)

	// Assert the result
	require.NoError(t, err)

	// Verify that the expected methods were called
	require.Equal(t, gameID, result.ID)
}
