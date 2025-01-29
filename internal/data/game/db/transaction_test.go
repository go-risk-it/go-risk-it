package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/mocks/internal_/data/game/pool"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestQueries_ExecuteInTransaction_ShouldRollbackIfPanic(t *testing.T) {
	t.Parallel()

	logContext := ctx.WithLog(context.Background(), zap.NewNop().Sugar())
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

	_, err := querier.ExecuteInTransaction(
		logContext,
		func(querier db.Querier) (interface{}, error) {
			panic("test")
		},
	)
	require.Error(t, err)

	mockDB.AssertExpectations(t)
	mockTransaction.AssertExpectations(t)
}

func TestQueries_ExecuteInTransaction_ShouldRollbackIfErr(t *testing.T) {
	t.Parallel()

	ctx := ctx.WithLog(context.Background(), zap.NewNop().Sugar())
	mockDB := pool.NewDB(t)
	mockTransaction := pool.NewTransaction(t)

	mockDB.EXPECT().
		BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).
		Return(mockTransaction, nil)
	mockTransaction.EXPECT().Rollback(ctx).Return(nil)

	querier := db.New(mockDB)

	_, err := querier.ExecuteInTransaction(ctx, func(querier db.Querier) (interface{}, error) {
		return nil, errors.New("test")
	})
	require.Error(t, err)

	mockDB.AssertExpectations(t)
	mockTransaction.AssertExpectations(t)
}

func TestQueries_ExecuteInTransaction_ShouldCommitIfNoErr(t *testing.T) {
	t.Parallel()

	ctx := ctx.WithLog(context.Background(), zap.NewNop().Sugar())
	mockDB := pool.NewDB(t)
	mockTransaction := pool.NewTransaction(t)

	mockDB.EXPECT().
		BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).
		Return(mockTransaction, nil)
	mockTransaction.EXPECT().Commit(ctx).Return(nil)

	querier := db.New(mockDB)

	_, err := querier.ExecuteInTransaction(ctx, func(querier db.Querier) (interface{}, error) {
		return -1, nil
	})
	require.NoError(t, err)

	mockDB.AssertExpectations(t)
	mockTransaction.AssertExpectations(t)
}
