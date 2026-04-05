package userpool_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/userpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPool() (*userpool.Pool, *atomic.Int64) {
	var counter atomic.Int64

	pool := userpool.New(userpool.Config{
		MaxConcurrentGames: 2,
		AuthFactory: func(_ context.Context) (*client.AuthResult, error) {
			id := counter.Add(1)

			return &client.AuthResult{
				UserID:      fakeUserID(id),
				AccessToken: fakeToken(id),
			}, nil
		},
	})

	return pool, &counter
}

func fakeUserID(id int64) string { return "user-" + itoa(id) }
func fakeToken(id int64) string  { return "token-" + itoa(id) }

func itoa(i int64) string {
	return []string{"", "1", "2", "3", "4", "5", "6", "7", "8"}[i]
}

func TestAcquire_LazyCreation(t *testing.T) {
	t.Parallel()

	pool, counter := newTestPool()

	entries, err := pool.Acquire(context.Background(), 4)
	require.NoError(t, err)
	require.Len(t, entries, 4)

	// Should have created exactly 4 users.
	assert.Equal(t, int64(4), counter.Load())

	// Each entry has a unique user ID.
	seen := make(map[string]bool)
	for _, e := range entries {
		assert.False(t, seen[e.Auth.UserID], "duplicate user ID: %s", e.Auth.UserID)
		seen[e.Auth.UserID] = true
	}
}

func TestAcquire_ReusesReleasedUsers(t *testing.T) {
	t.Parallel()

	pool, counter := newTestPool()

	// Acquire 4 users.
	entries1, err := pool.Acquire(context.Background(), 4)
	require.NoError(t, err)
	assert.Equal(t, int64(4), counter.Load())

	// Release all.
	pool.Release(entries1)

	// Acquire 4 again — should reuse, not create new.
	entries2, err := pool.Acquire(context.Background(), 4)
	require.NoError(t, err)
	assert.Equal(t, int64(4), counter.Load(), "should not create new users")

	// Same user IDs returned.
	ids1 := make(map[string]bool)
	for _, e := range entries1 {
		ids1[e.Auth.UserID] = true
	}

	for _, e := range entries2 {
		assert.True(t, ids1[e.Auth.UserID], "expected reused user, got new: %s", e.Auth.UserID)
	}
}

func TestAcquire_ConcurrencyLimit(t *testing.T) {
	t.Parallel()

	pool, counter := newTestPool()

	// Acquire 2 users (each at 1 active game).
	entries1, err := pool.Acquire(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), counter.Load())

	// Acquire 2 more — same users should be reused (capacity=2, currently at 1).
	entries2, err := pool.Acquire(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), counter.Load(), "should reuse users at capacity 1/2")

	// Now all 2 users are at capacity (2/2). Acquire 2 more — must create new.
	entries3, err := pool.Acquire(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, int64(4), counter.Load(), "should create 2 new users")

	pool.Release(entries1)
	pool.Release(entries2)
	pool.Release(entries3)
}

func TestAcquire_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Already cancelled.

	pool := userpool.New(userpool.Config{
		MaxConcurrentGames: 2,
		AuthFactory: func(_ context.Context) (*client.AuthResult, error) {
			return &client.AuthResult{UserID: "u", AccessToken: "t"}, nil
		},
	})

	// No existing users, needs to create → should fail on cancelled context.
	entries, err := pool.Acquire(ctx, 2)
	require.Error(t, err)
	assert.Nil(t, entries)
}

func TestAcquire_AuthFactoryError(t *testing.T) {
	t.Parallel()

	pool := userpool.New(userpool.Config{
		MaxConcurrentGames: 2,
		AuthFactory: func(_ context.Context) (*client.AuthResult, error) {
			return nil, errors.New("auth down")
		},
	})

	entries, err := pool.Acquire(context.Background(), 2)
	require.Error(t, err)
	assert.Nil(t, entries)
	assert.Contains(t, err.Error(), "auth down")
}

func TestRelease_NilSafe(t *testing.T) {
	t.Parallel()

	pool, _ := newTestPool()

	// Should not panic.
	pool.Release(nil)
	pool.Release([]*userpool.Entry{})
}

func TestConcurrentAcquireRelease(t *testing.T) {
	t.Parallel()

	var counter atomic.Int64

	pool := userpool.New(userpool.Config{
		MaxConcurrentGames: 2,
		AuthFactory: func(_ context.Context) (*client.AuthResult, error) {
			id := counter.Add(1)

			return &client.AuthResult{
				UserID:      fmt.Sprintf("u%d", id),
				AccessToken: fmt.Sprintf("t%d", id),
			}, nil
		},
	})

	var wg sync.WaitGroup

	for range 20 {
		wg.Go(func() {
			entries, err := pool.Acquire(context.Background(), 2)
			if err != nil {
				return
			}

			// Simulate game.
			pool.Release(entries)
		})
	}

	wg.Wait()

	stats := pool.Stats()
	assert.Equal(t, 0, stats.ActiveSlots, "all slots should be released")
	assert.Positive(t, stats.TotalUsers)
}
