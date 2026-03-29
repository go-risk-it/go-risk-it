package start_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/start"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/lobby/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (*db.Querier, start.Service) {
	t.Helper()

	querier := db.NewQuerier(t)
	service := start.NewService(querier)

	return querier, service
}

func lobbyContext() ctx.LobbyContext {
	userID := "giovanni"

	traceCtx := ctx.WithSpan(
		context.Background(),
		noop.Span{},
	)
	userCtx := ctx.WithUserID(traceCtx, userID)

	return ctx.WithLobbyID(userCtx, int64(42))
}

func TestServiceImpl_CanStartLobby(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		canStart bool
	}

	tests := []testCase{
		{
			name:     "When lobby can be started",
			canStart: true,
		},
		{
			name:     "When lobby cannot be started",
			canStart: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, service := setup(t)
			lctx := lobbyContext()

			querier.
				EXPECT().
				CanLobbyBeStarted(lctx, sqlc.CanLobbyBeStartedParams{
					LobbyID:             int64(42),
					UserID:              "giovanni",
					MinimumParticipants: 3,
				}).
				Return(test.canStart, nil)

			result, err := service.CanStartLobby(lctx)

			require.NoError(t, err)
			require.Equal(t, test.canStart, result)
		})
	}
}

func TestServiceImpl_CanStartLobby_Error(t *testing.T) {
	t.Parallel()

	querier, service := setup(t)
	lctx := lobbyContext()

	querier.
		EXPECT().
		CanLobbyBeStarted(lctx, sqlc.CanLobbyBeStartedParams{
			LobbyID:             int64(42),
			UserID:              "giovanni",
			MinimumParticipants: 3,
		}).
		Return(false, errors.New("db error"))

	result, err := service.CanStartLobby(lctx)

	require.Error(t, err)
	require.EqualError(t, err, "failed to check if lobby can be started: db error")
	require.False(t, result)
}

func TestServiceImpl_GetLobbyPlayers_Success(t *testing.T) {
	t.Parallel()

	querier, service := setup(t)
	lctx := lobbyContext()

	players := []sqlc.GetLobbyPlayersRow{
		{UserID: "giovanni", Name: "Giovanni"},
		{UserID: "gabriele", Name: "Gabriele"},
		{UserID: "marco", Name: "Marco"},
	}

	querier.
		EXPECT().
		GetLobbyPlayers(lctx, int64(42)).
		Return(players, nil)

	result, err := service.GetLobbyPlayers(lctx)

	require.NoError(t, err)
	require.Equal(t, players, result)
}

func TestServiceImpl_GetLobbyPlayers_Error(t *testing.T) {
	t.Parallel()

	querier, service := setup(t)
	lctx := lobbyContext()

	querier.
		EXPECT().
		GetLobbyPlayers(lctx, int64(42)).
		Return(nil, errors.New("db error"))

	result, err := service.GetLobbyPlayers(lctx)

	require.Error(t, err)
	require.EqualError(t, err, "failed to get lobby players: db error")
	require.Nil(t, result)
}

func TestServiceImpl_MarkLobbyAsStarted_Success(t *testing.T) {
	t.Parallel()

	querier, service := setup(t)
	lctx := lobbyContext()

	gameID := int64(99)

	querier.
		EXPECT().
		MarkLobbyAsStarted(lctx, sqlc.MarkLobbyAsStartedParams{
			LobbyID: int64(42),
			GameID: pgtype.Int8{
				Int64: gameID,
				Valid: true,
			},
		}).
		Return(nil)

	err := service.MarkLobbyAsStarted(lctx, gameID)

	require.NoError(t, err)
}

func TestServiceImpl_MarkLobbyAsStarted_Error(t *testing.T) {
	t.Parallel()

	querier, service := setup(t)
	lctx := lobbyContext()

	gameID := int64(99)

	querier.
		EXPECT().
		MarkLobbyAsStarted(lctx, sqlc.MarkLobbyAsStartedParams{
			LobbyID: int64(42),
			GameID: pgtype.Int8{
				Int64: gameID,
				Valid: true,
			},
		}).
		Return(errors.New("update error"))

	err := service.MarkLobbyAsStarted(lctx, gameID)

	require.Error(t, err)
	require.EqualError(t, err, "failed to mark lobby as started: update error")
}
