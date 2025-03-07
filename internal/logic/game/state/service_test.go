package state_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

func TestServiceImpl_GetGameState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	querier := db.NewQuerier(t)

	// Initialize the state under test
	service := state.NewService(querier)

	// Set up test data
	gameID := int64(1)
	ctx := ctx.WithGameID(
		ctx.WithUserID(
			ctx.WithSpan(ctx.WithLog(context.Background(), zap.NewExample().Sugar()), noop.Span{}),
			"francesco",
		),
		gameID,
	)

	// Set up expectations for GetGame method
	querier.EXPECT().GetGame(ctx, gameID).Return(sqlc.GetGameRow{
		ID:           gameID,
		CurrentPhase: sqlc.GamePhaseTypeATTACK,
		Turn:         3,
		WinnerUserID: pgtype.Text{
			Valid:  false,
			String: "",
		},
	}, nil)

	// Call the method under test
	result, err := service.GetGameState(ctx)

	// Assert the result
	require.NoError(t, err)

	// Verify that the expected methods were called
	require.Equal(t, gameID, result.ID)
	require.Equal(t, "", result.WinnerUserID)
}
