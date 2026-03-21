package db_test

import (
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/lobby/pool"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInTransaction_ShouldRollbackIfPanic(t *testing.T) {
	t.Parallel()

	logContext := ctx.WithLog(t.Context(), zap.NewNop().Sugar())
	mockDB := pool.NewDB(t)
	mockTransaction := pool.NewTransaction(t)

	mockDB.EXPECT().
		BeginTx(logContext, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).
		Return(mockTransaction, nil)
	mockTransaction.EXPECT().Rollback(logContext).Return(nil)

	querier := db.New(mockDB)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	_, err := dbutil.InTransaction(
		querier,
		logContext,
		func(querier db.Querier) (any, error) {
			panic("test")
		},
	)
	require.Error(t, err)

	mockDB.AssertExpectations(t)
	mockTransaction.AssertExpectations(t)
}

func TestInTransaction_ShouldRollbackIfErr(t *testing.T) {
	t.Parallel()

	ctx := ctx.WithLog(t.Context(), zap.NewNop().Sugar())
	mockDB := pool.NewDB(t)
	mockTransaction := pool.NewTransaction(t)

	mockDB.EXPECT().
		BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).
		Return(mockTransaction, nil)
	mockTransaction.EXPECT().Rollback(ctx).Return(nil)

	querier := db.New(mockDB)

	_, err := dbutil.InTransaction(querier, ctx, func(querier db.Querier) (any, error) {
		return nil, errors.New("test")
	})
	require.Error(t, err)

	mockDB.AssertExpectations(t)
	mockTransaction.AssertExpectations(t)
}

func TestInTransaction_ShouldCommitIfNoErr(t *testing.T) {
	t.Parallel()

	ctx := ctx.WithLog(t.Context(), zap.NewNop().Sugar())
	mockDB := pool.NewDB(t)
	mockTransaction := pool.NewTransaction(t)

	mockDB.EXPECT().
		BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).
		Return(mockTransaction, nil)
	mockTransaction.EXPECT().Commit(ctx).Return(nil)

	querier := db.New(mockDB)

	_, err := dbutil.InTransaction(querier, ctx, func(querier db.Querier) (any, error) {
		return -1, nil
	})
	require.NoError(t, err)

	mockDB.AssertExpectations(t)
	mockTransaction.AssertExpectations(t)
}
