package state_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/data/db"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestService_GetGameState(t *testing.T) {
	t.Parallel()

	// Initialize dependencies
	querier := db.NewQuerier(t)

	// Initialize the service under test
	service := state.NewService(querier)

	// Set up test data
	gameID := int64(1)
	ctx := ctx.WithGameID(
		kernelctx.WithUserID(
			kernelctx.WithSpan(t.Context(), noop.Span{}),
			"francesco",
		),
		gameID,
	)

	// Set up expectations for GetGame method
	querier.EXPECT().GetGame(mock.Anything, gameID).Return(sqlc.GetGameRow{
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
	require.Empty(t, result.WinnerUserID)
}
