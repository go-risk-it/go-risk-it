package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockTx implements pgx.Tx for testing. Only Commit/Rollback are exercised.
type mockTx struct {
	mock.Mock
}

func (m *mockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)

	return args.Get(0).(pgx.Tx), args.Error(1) //nolint:forcetypeassert // test mock
}

func (m *mockTx) Commit(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockTx) CopyFrom(
	_ context.Context,
	_ pgx.Identifier,
	_ []string,
	_ pgx.CopyFromSource,
) (int64, error) {
	return 0, nil
}

func (m *mockTx) SendBatch(_ context.Context, _ *pgx.Batch) pgx.BatchResults {
	return nil
}

func (m *mockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (m *mockTx) Prepare(
	_ context.Context,
	_ string,
	_ string,
) (*pgconn.StatementDescription, error) {
	return nil, nil //nolint:nilnil // test stub
}

func (m *mockTx) Exec(
	_ context.Context,
	_ string,
	_ ...any,
) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (m *mockTx) Query(
	_ context.Context,
	_ string,
	_ ...any,
) (pgx.Rows, error) {
	return nil, nil //nolint:nilnil // test stub
}

func (m *mockTx) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return nil
}

func (m *mockTx) Conn() *pgx.Conn {
	return nil
}

// mockQuerier is a minimal Transactable implementation for testing InTransaction.
type mockQuerier struct {
	mock.Mock
}

func (m *mockQuerier) BeginTx(
	ctx context.Context,
	txOptions pgx.TxOptions,
) (pgx.Tx, error) {
	args := m.Called(ctx, txOptions)

	return args.Get(0).(pgx.Tx), args.Error(1) //nolint:forcetypeassert // test mock
}

func (m *mockQuerier) WithTx(_ pgx.Tx) *mockQuerier {
	return m
}

var _ db.Transactable[*mockQuerier] = (*mockQuerier)(nil)

func TestInTransaction_ShouldRollbackIfPanic(t *testing.T) {
	t.Parallel()

	logContext := ctx.WithLog(t.Context(), zap.NewNop().Sugar())
	querier := &mockQuerier{}
	transaction := &mockTx{}

	querier.On("BeginTx", logContext, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).
		Return(transaction, nil)
	transaction.On("Rollback", logContext).Return(nil)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	_, err := db.InTransaction(
		querier,
		logContext,
		nil,
		func(_ *mockQuerier) (any, error) {
			panic("test")
		},
	)
	require.Error(t, err)

	querier.AssertExpectations(t)
	transaction.AssertExpectations(t)
}

func TestInTransaction_ShouldRollbackIfErr(t *testing.T) {
	t.Parallel()

	logContext := ctx.WithLog(t.Context(), zap.NewNop().Sugar())
	querier := &mockQuerier{}
	transaction := &mockTx{}

	querier.On("BeginTx", logContext, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).
		Return(transaction, nil)
	transaction.On("Rollback", logContext).Return(nil)

	_, err := db.InTransaction(querier, logContext, nil, func(_ *mockQuerier) (any, error) {
		return nil, errors.New("test")
	})
	require.Error(t, err)

	querier.AssertExpectations(t)
	transaction.AssertExpectations(t)
}

func TestInTransaction_ShouldCommitIfNoErr(t *testing.T) {
	t.Parallel()

	logContext := ctx.WithLog(t.Context(), zap.NewNop().Sugar())
	querier := &mockQuerier{}
	transaction := &mockTx{}

	querier.On("BeginTx", logContext, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).
		Return(transaction, nil)
	transaction.On("Commit", logContext).Return(nil)

	_, err := db.InTransaction(querier, logContext, nil, func(_ *mockQuerier) (any, error) {
		return -1, nil
	})
	require.NoError(t, err)

	querier.AssertExpectations(t)
	transaction.AssertExpectations(t)
}
