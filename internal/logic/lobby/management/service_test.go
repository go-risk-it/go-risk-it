package management_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/management"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/logic/lobby/signals"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

func setup(t *testing.T) (
	*db.Querier,
	*management.ServiceImpl,
) {
	t.Helper()

	querier := db.NewQuerier(t)
	signal := signals.NewLobbyStateChangedSignal(t)
	service := management.NewService(querier, signal)

	return querier, service
}

func userContext() ctx.UserContext {
	userID := "giovanni"

	traceCtx := ctx.WithSpan(
		ctx.WithLog(context.Background(), zap.NewExample().Sugar()),
		noop.Span{},
	)

	return ctx.WithUserID(traceCtx, userID)
}

func lobbyContext() ctx.LobbyContext {
	return ctx.WithLobbyID(userContext(), int64(42))
}

func TestServiceImpl_JoinLobbyQ_Success(t *testing.T) {
	t.Parallel()

	querier, service := setup(t)
	lctx := lobbyContext()

	querier.
		EXPECT().
		InsertParticipant(lctx, sqlc.InsertParticipantParams{
			LobbyID: int64(42),
			UserID:  "giovanni",
			Name:    "Giovanni",
		}).
		Return(int64(7), nil)

	err := service.JoinLobbyQ(lctx, querier, "Giovanni")

	require.NoError(t, err)
}

func TestServiceImpl_JoinLobbyQ_InsertFails(t *testing.T) {
	t.Parallel()

	querier, service := setup(t)
	lctx := lobbyContext()

	querier.
		EXPECT().
		InsertParticipant(lctx, sqlc.InsertParticipantParams{
			LobbyID: int64(42),
			UserID:  "giovanni",
			Name:    "Giovanni",
		}).
		Return(int64(0), errors.New("duplicate participant"))

	err := service.JoinLobbyQ(lctx, querier, "Giovanni")

	require.Error(t, err)
	require.EqualError(t, err, "failed to insert participant: duplicate participant")
}

func TestServiceImpl_GetUserLobbiesQ_Success(t *testing.T) {
	t.Parallel()

	querier, service := setup(t)
	uctx := userContext()

	ownedLobbies := []sqlc.GetOwnedLobbiesRow{
		{ID: 1, GameID: pgtype.Int8{}, ParticipantCount: 3},
	}
	joinedLobbies := []sqlc.GetJoinedLobbiesRow{
		{ID: 2, GameID: pgtype.Int8{}, ParticipantCount: 4},
	}
	joinableLobbies := []sqlc.GetJoinableLobbiesRow{
		{ID: 3, GameID: pgtype.Int8{}, ParticipantCount: 2},
	}

	querier.
		EXPECT().
		GetOwnedLobbies(uctx, "giovanni").
		Return(ownedLobbies, nil)
	querier.
		EXPECT().
		GetJoinedLobbies(uctx, "giovanni").
		Return(joinedLobbies, nil)
	querier.
		EXPECT().
		GetJoinableLobbies(uctx, "giovanni").
		Return(joinableLobbies, nil)

	result, err := service.GetUserLobbiesQ(uctx, querier)

	require.NoError(t, err)
	require.Equal(t, ownedLobbies, result.Owned)
	require.Equal(t, joinedLobbies, result.Joined)
	require.Equal(t, joinableLobbies, result.Joinable)
}

func TestServiceImpl_GetUserLobbiesQ_Failures(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		setupMocks    func(*db.Querier, ctx.UserContext)
		expectedError string
	}

	tests := []testCase{
		{
			name: "When GetOwnedLobbies fails",
			setupMocks: func(querier *db.Querier, uctx ctx.UserContext) {
				querier.
					EXPECT().
					GetOwnedLobbies(uctx, "giovanni").
					Return(nil, errors.New("owned query error"))
			},
			expectedError: "failed to get owned lobbies: owned query error",
		},
		{
			name: "When GetJoinedLobbies fails",
			setupMocks: func(querier *db.Querier, uctx ctx.UserContext) {
				querier.
					EXPECT().
					GetOwnedLobbies(uctx, "giovanni").
					Return([]sqlc.GetOwnedLobbiesRow{}, nil)
				querier.
					EXPECT().
					GetJoinedLobbies(uctx, "giovanni").
					Return(nil, errors.New("joined query error"))
			},
			expectedError: "failed to get joined lobbies: joined query error",
		},
		{
			name: "When GetJoinableLobbies fails",
			setupMocks: func(querier *db.Querier, uctx ctx.UserContext) {
				querier.
					EXPECT().
					GetOwnedLobbies(uctx, "giovanni").
					Return([]sqlc.GetOwnedLobbiesRow{}, nil)
				querier.
					EXPECT().
					GetJoinedLobbies(uctx, "giovanni").
					Return([]sqlc.GetJoinedLobbiesRow{}, nil)
				querier.
					EXPECT().
					GetJoinableLobbies(uctx, "giovanni").
					Return(nil, errors.New("joinable query error"))
			},
			expectedError: "failed to get joinable lobbies: joinable query error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			querier, service := setup(t)
			uctx := userContext()

			test.setupMocks(querier, uctx)

			result, err := service.GetUserLobbiesQ(uctx, querier)

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)
			require.Nil(t, result)
		})
	}
}
