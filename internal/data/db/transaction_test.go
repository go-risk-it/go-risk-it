package db_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomfran/go-risk-it/internal/data/db"
	"github.com/tomfran/go-risk-it/mocks/internal_/data/pool"
	"go.uber.org/zap"
)

func TestQueries_ExecuteInTransaction_ShouldRollbackIfPanic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockDB := pool.NewDB(t)
	mockTransaction := pool.NewTransaction(t)

	mockDB.EXPECT().Begin(ctx).Return(mockTransaction, nil)
	mockTransaction.EXPECT().Rollback(ctx).Return(nil)

	querier := db.New(mockDB, zap.NewNop().Sugar())

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	err := querier.ExecuteInTransaction(ctx, func(querier db.Querier) error {
		panic("test")
	})
	require.Error(t, err)

	mockDB.AssertExpectations(t)
	mockTransaction.AssertExpectations(t)
}

func TestQueries_ExecuteInTransaction_ShouldRollbackIfErr(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockDB := pool.NewDB(t)
	mockTransaction := pool.NewTransaction(t)

	mockDB.EXPECT().Begin(ctx).Return(mockTransaction, nil)
	mockTransaction.EXPECT().Rollback(ctx).Return(nil)

	querier := db.New(mockDB, zap.NewNop().Sugar())

	err := querier.ExecuteInTransaction(ctx, func(querier db.Querier) error {
		return fmt.Errorf("test")
	})
	require.Error(t, err)

	mockDB.AssertExpectations(t)
	mockTransaction.AssertExpectations(t)
}

func TestQueries_ExecuteInTransaction_ShouldCommitIfNoErr(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockDB := pool.NewDB(t)
	mockTransaction := pool.NewTransaction(t)

	mockDB.EXPECT().Begin(ctx).Return(mockTransaction, nil)
	mockTransaction.EXPECT().Commit(ctx).Return(nil)

	querier := db.New(mockDB, zap.NewNop().Sugar())

	err := querier.ExecuteInTransaction(ctx, func(querier db.Querier) error {
		return nil
	})
	require.NoError(t, err)

	mockDB.AssertExpectations(t)
	mockTransaction.AssertExpectations(t)
}
