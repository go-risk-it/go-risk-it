package orchestration_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/phase"
	dbmock "github.com/go-risk-it/go-risk-it/internal/game/testmocks/data/db"
	phasemock "github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/phase"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// setupPersistTx sets up a mock querier to support InTransactionWithIsolation with ReadCommitted.
func setupPersistTx(t *testing.T, querier *dbmock.Querier) *dbmock.Querier {
	t.Helper()

	transaction := dbmock.NewTransaction(t)
	txQuerier := dbmock.NewQuerier(t)

	querier.EXPECT().
		BeginTx(mock.Anything, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).
		Return(transaction, nil)

	querier.EXPECT().
		WithTx(transaction).
		Return(txQuerier)

	transaction.EXPECT().
		Commit(mock.Anything).
		Return(nil)

	return txQuerier
}

// setupPersistTxWithRollback sets up a mock that expects rollback (for error paths).
func setupPersistTxWithRollback(t *testing.T, querier *dbmock.Querier) *dbmock.Querier {
	t.Helper()

	transaction := dbmock.NewTransaction(t)
	txQuerier := dbmock.NewQuerier(t)

	querier.EXPECT().
		BeginTx(mock.Anything, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).
		Return(transaction, nil)

	querier.EXPECT().
		WithTx(transaction).
		Return(txQuerier)

	transaction.EXPECT().
		Rollback(mock.Anything).
		Return(nil)

	return txQuerier
}

func persistCtx() gamectx.GameContext {
	traceCtx := kernelctx.WithSpan(context.Background(), tracenoop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "user1")

	return gamectx.WithGameID(userCtx, 123)
}

func TestPersist_EmptyEffect(t *testing.T) {
	t.Parallel()

	effect := &orchestration.PersistenceEffect{}
	querier := dbmock.NewQuerier(t)
	txQuerier := setupPersistTx(t, querier)
	_ = txQuerier // no writes expected

	mockPhaseService := phasemock.NewService(t)
	gameCtx := persistCtx()

	err := orchestration.Persist(gameCtx, querier, mockPhaseService, effect)
	require.NoError(t, err)
}

func TestPersist_MoveLogOnly(t *testing.T) {
	t.Parallel()

	effect := &orchestration.PersistenceEffect{
		MoveLog: &orchestration.MoveLogEntry{
			GameID:    123,
			UserID:    "user1",
			PhaseType: "ATTACK",
			MoveData:  []byte(`{"type":"attack"}`),
			Result:    []byte(`{"success":true}`),
		},
	}

	querier := dbmock.NewQuerier(t)
	txQuerier := setupPersistTx(t, querier)

	txQuerier.EXPECT().
		CreateMoveLog(mock.Anything, sqlc.CreateMoveLogParams{
			GameID:   123,
			UserID:   "user1",
			Phase:    sqlc.GamePhaseType("ATTACK"),
			MoveData: []byte(`{"type":"attack"}`),
			Result:   []byte(`{"success":true}`),
		}).
		Return(sqlc.GameMoveLog{ID: 1}, nil).
		Once()

	mockPhaseService := phasemock.NewService(t)
	gameCtx := persistCtx()

	err := orchestration.Persist(gameCtx, querier, mockPhaseService, effect)
	require.NoError(t, err)
}

func TestPersist_PhaseTransitionWithConquer(t *testing.T) {
	t.Parallel()

	effect := &orchestration.PersistenceEffect{
		PhaseTransition: &orchestration.PhaseTransition{
			Turn:             1,
			PhaseType:        "CONQUER",
			CurrentPhaseType: "ATTACK",
			Players: []orchestration.PlayerRef{
				{UserID: "user1", InternalID: 10},
			},
			ConquerData: &orchestration.ConquerData{
				SourceRegionName: "alaska",
				TargetRegionName: "kamchatka",
				MinTroops:        1,
			},
		},
	}

	querier := dbmock.NewQuerier(t)
	txQuerier := setupPersistTx(t, querier)

	mockPhaseService := phasemock.NewService(t)
	phaseID := int64(999)

	mockPhaseService.EXPECT().
		InsertPhase(mock.Anything, txQuerier, phase.PhaseInsertParams{
			PhaseType:    sqlc.GamePhaseType("CONQUER"),
			CurrentPhase: sqlc.GamePhaseType("ATTACK"),
			Turn:         1,
			Players: []snapshot.PlayerState{
				{UserID: "user1"},
			},
		}).
		Return(&sqlc.GamePhase{ID: phaseID}, nil).
		Once()

	txQuerier.EXPECT().
		InsertConquerPhase(mock.Anything, mock.MatchedBy(func(params sqlc.InsertConquerPhaseParams) bool {
			return params.PhaseID == phaseID
		})).
		Return(sqlc.GameConquerPhase{}, nil).
		Once()

	gameCtx := persistCtx()

	err := orchestration.Persist(gameCtx, querier, mockPhaseService, effect)
	require.NoError(t, err)
}

func TestPersist_PhaseTransitionWithDeploy(t *testing.T) {
	t.Parallel()

	effect := &orchestration.PersistenceEffect{
		PhaseTransition: &orchestration.PhaseTransition{
			Turn:             1,
			PhaseType:        "DEPLOY",
			CurrentPhaseType: "REINFORCE",
			Players: []orchestration.PlayerRef{
				{UserID: "user1", InternalID: 10},
			},
			DeployData: &orchestration.DeployData{
				DeployableTroops: 5,
			},
		},
	}

	querier := dbmock.NewQuerier(t)
	txQuerier := setupPersistTx(t, querier)

	mockPhaseService := phasemock.NewService(t)
	phaseID := int64(888)

	mockPhaseService.EXPECT().
		InsertPhase(mock.Anything, txQuerier, mock.Anything).
		Return(&sqlc.GamePhase{ID: phaseID}, nil).
		Once()

	txQuerier.EXPECT().
		InsertDeployPhase(mock.Anything, sqlc.InsertDeployPhaseParams{
			PhaseID:          phaseID,
			DeployableTroops: 5,
		}).
		Return(sqlc.GameDeployPhase{}, nil).
		Once()

	gameCtx := persistCtx()

	err := orchestration.Persist(gameCtx, querier, mockPhaseService, effect)
	require.NoError(t, err)
}

func TestPersist_EliminationCascade(t *testing.T) {
	t.Parallel()

	effect := &orchestration.PersistenceEffect{
		Elimination: &orchestration.EliminationEffect{
			EliminatedUserID: "user2",
			ConquerorUserID:  "user1",
		},
	}

	querier := dbmock.NewQuerier(t)
	txQuerier := setupPersistTx(t, querier)

	txQuerier.EXPECT().
		GetPlayerByUserId(mock.Anything, "user2").
		Return(sqlc.GamePlayer{ID: 20}, nil).
		Once()

	txQuerier.EXPECT().
		TransferCardsOwnership(mock.Anything, sqlc.TransferCardsOwnershipParams{
			GameID: 123,
			To:     "user1",
			From:   pgtype.Int8{Int64: 20, Valid: true},
		}).
		Return(nil).
		Once()

	txQuerier.EXPECT().
		ReassignMissions(mock.Anything, sqlc.ReassignMissionsParams{
			GameID:             123,
			UserID:             "user2",
			EliminatedPlayerID: 20,
		}).
		Return(nil).
		Once()

	txQuerier.EXPECT().
		DeleteSpuriousEliminatePlayerMissions(mock.Anything, int64(123)).
		Return(nil).
		Once()

	mockPhaseService := phasemock.NewService(t)
	gameCtx := persistCtx()

	err := orchestration.Persist(gameCtx, querier, mockPhaseService, effect)
	require.NoError(t, err)
}

func TestPersist_ErrorRollback(t *testing.T) {
	t.Parallel()

	effect := &orchestration.PersistenceEffect{
		MoveLog: &orchestration.MoveLogEntry{
			GameID:    123,
			UserID:    "user1",
			PhaseType: "ATTACK",
			MoveData:  []byte(`{}`),
			Result:    []byte(`{}`),
		},
		MoveExecution: &orchestration.MoveExecution{
			RegionTroopUpdates: []orchestration.RegionTroopUpdate{
				{RegionID: 1, Delta: -3},
				{RegionID: 2, Delta: 2},
			},
		},
	}

	querier := dbmock.NewQuerier(t)
	txQuerier := setupPersistTxWithRollback(t, querier)

	txQuerier.EXPECT().
		CreateMoveLog(mock.Anything, mock.Anything).
		Return(sqlc.GameMoveLog{ID: 1}, nil).
		Once()

	txQuerier.EXPECT().
		IncreaseRegionTroops(mock.Anything, sqlc.IncreaseRegionTroopsParams{ID: 1, Troops: -3}).
		Return(nil).
		Once()

	expectedErr := domainerrors.NewValidationError("region not found")
	txQuerier.EXPECT().
		IncreaseRegionTroops(mock.Anything, sqlc.IncreaseRegionTroopsParams{ID: 2, Troops: 2}).
		Return(expectedErr).
		Once()

	mockPhaseService := phasemock.NewService(t)
	gameCtx := persistCtx()

	err := orchestration.Persist(gameCtx, querier, mockPhaseService, effect)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "region not found")
}

func TestPersist_GameConclusion(t *testing.T) {
	t.Parallel()

	effect := &orchestration.PersistenceEffect{
		GameConclusion: &orchestration.GameConclusion{
			WinnerUserID: "user1",
		},
	}

	querier := dbmock.NewQuerier(t)
	txQuerier := setupPersistTx(t, querier)

	txQuerier.EXPECT().
		GetPlayerByUserId(mock.Anything, "user1").
		Return(sqlc.GamePlayer{ID: 10}, nil).
		Once()

	txQuerier.EXPECT().
		AssignGameWinner(mock.Anything, sqlc.AssignGameWinnerParams{
			WinnerPlayerID: pgtype.Int8{Int64: 10, Valid: true},
			GameID:         123,
		}).
		Return(nil).
		Once()

	mockPhaseService := phasemock.NewService(t)
	gameCtx := persistCtx()

	err := orchestration.Persist(gameCtx, querier, mockPhaseService, effect)
	require.NoError(t, err)
}
