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
				UserID:      fmt.Sprintf("user-%d", id),
				AccessToken: fmt.Sprintf("token-%d", id),
			}, nil
		},
	})

	return pool, &counter
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

// TestGameReplacementPattern simulates the staircase game replacement cycle:
// maintain N concurrent games, each acquiring 4 users. When a game finishes
// (release 4), a replacement immediately acquires 4. Pool should stabilize
// at N*4/maxConcurrent users and stop growing.
func TestGameReplacementPattern(t *testing.T) {
	t.Parallel()

	pool, counter := newTestPool()
	const concurrentGames = 10
	const playersPerGame = 4
	const replacements = 50

	// Initial fill: 10 games × 4 players = 40 users (maxConcurrent=2, so 20 users).
	activeGames := make([][]*userpool.Entry, concurrentGames)

	for i := range concurrentGames {
		entries, err := pool.Acquire(context.Background(), playersPerGame)
		require.NoError(t, err)
		activeGames[i] = entries
	}

	usersAfterFill := counter.Load()

	// Replacement cycle: finish game[i%10], start a new one.
	for i := range replacements {
		slot := i % concurrentGames
		pool.Release(activeGames[slot])

		entries, err := pool.Acquire(context.Background(), playersPerGame)
		require.NoError(t, err)
		activeGames[slot] = entries
	}

	usersAfterReplacements := counter.Load()

	// Pool should NOT have grown during replacements — all users reused.
	assert.Equal(
		t,
		usersAfterFill,
		usersAfterReplacements,
		"pool should not grow during replacement cycle (created %d users for fill, %d after %d replacements)",
		usersAfterFill,
		usersAfterReplacements,
		replacements,
	)

	// Cleanup.
	for _, entries := range activeGames {
		pool.Release(entries)
	}

	stats := pool.Stats()
	assert.Equal(t, 0, stats.ActiveSlots)
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
