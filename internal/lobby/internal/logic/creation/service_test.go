package creation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic/creation"
	"github.com/go-risk-it/go-risk-it/internal/lobby/testmocks/data/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func setup(t *testing.T) (*db.Querier, creation.Service) {
	t.Helper()

	querier := db.NewQuerier(t)
	service := creation.NewService(querier, nil)

	return querier, service
}

func userContext() ctx.UserContext {
	userID := "giovanni"

	traceCtx := ctx.WithSpan(
		context.Background(),
		noop.Span{},
	)

	return ctx.WithUserID(traceCtx, userID)
}

func TestServiceImpl_CreateLobbyWithQuerier_Success(t *testing.T) {
	t.Parallel()

	querier, service := setup(t)
	uctx := userContext()

	lobbyID := int64(42)
	participantID := int64(7)

	querier.
		EXPECT().
		CreateLobby(uctx).
		Return(lobbyID, nil)

	querier.
		EXPECT().
		InsertParticipant(uctx, sqlc.InsertParticipantParams{
			LobbyID: lobbyID,
			UserID:  "giovanni",
			Name:    "Giovanni",
		}).
		Return(participantID, nil)

	querier.
		EXPECT().
		UpdateLobbyOwner(uctx, sqlc.UpdateLobbyOwnerParams{
			OwnerID: pgtype.Int8{
				Int64: participantID,
				Valid: true,
			},
			ID: lobbyID,
		}).
		Return(nil)

	result, err := service.CreateLobbyWithQuerier(uctx, querier, "Giovanni")

	require.NoError(t, err)
	require.Equal(t, lobbyID, result)
}

func TestServiceImpl_CreateLobbyWithQuerier_Failures(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		setupMocks    func(*db.Querier, ctx.UserContext)
		expectedError string
	}

	tests := []testCase{
		{
			name: "When CreateLobby fails",
			setupMocks: func(querier *db.Querier, uctx ctx.UserContext) {
				querier.
					EXPECT().
					CreateLobby(uctx).
					Return(int64(0), errors.New("db error"))
			},
			expectedError: "failed to create lobby: db error",
		},
		{
			name: "When InsertParticipant fails",
			setupMocks: func(querier *db.Querier, uctx ctx.UserContext) {
				querier.
					EXPECT().
					CreateLobby(uctx).
					Return(int64(42), nil)

				querier.
					EXPECT().
					InsertParticipant(uctx, sqlc.InsertParticipantParams{
						LobbyID: 42,
						UserID:  "giovanni",
						Name:    "Giovanni",
					}).
					Return(int64(0), errors.New("insert error"))
			},
			expectedError: "failed to insert participant: insert error",
		},
		{
			name: "When UpdateLobbyOwner fails",
			setupMocks: func(querier *db.Querier, uctx ctx.UserContext) {
				querier.
					EXPECT().
					CreateLobby(uctx).
					Return(int64(42), nil)

				querier.
					EXPECT().
					InsertParticipant(uctx, sqlc.InsertParticipantParams{
						LobbyID: 42,
						UserID:  "giovanni",
						Name:    "Giovanni",
					}).
					Return(int64(7), nil)

				querier.
					EXPECT().
					UpdateLobbyOwner(uctx, sqlc.UpdateLobbyOwnerParams{
						OwnerID: pgtype.Int8{
							Int64: 7,
							Valid: true,
						},
						ID: 42,
					}).
					Return(errors.New("update error"))
			},
			expectedError: "failed to update lobby owner: update error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, service := setup(t)
			uctx := userContext()

			test.setupMocks(querier, uctx)

			result, err := service.CreateLobbyWithQuerier(uctx, querier, "Giovanni")

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
			require.Equal(t, int64(-1), result)
		})
	}
}
