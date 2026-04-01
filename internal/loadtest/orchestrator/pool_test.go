package orchestrator_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_MaintainsConcurrency(t *testing.T) {
	t.Parallel()

	var launched atomic.Int64

	fake := func(ctx context.Context, idx, players int) orchestrator.GameResult {
		launched.Add(1)
		time.Sleep(50 * time.Millisecond)

		return orchestrator.GameResult{GameIndex: idx}
	}

	pool := orchestrator.NewPool(
		orchestrator.PoolConfig{
			TargetGames:  3,
			NumPlayers:   4,
			StaggerDelay: 10 * time.Millisecond,
		},
		fake,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	go pool.Run(ctx)
	<-ctx.Done()
	pool.WaitDrain()

	// 3 concurrent × ~50ms per game × 300ms = ~18 total, but with stagger and
	// scheduling jitter, at least 6 should have launched.
	assert.GreaterOrEqual(t, launched.Load(), int64(6))
}

func TestPool_ReadySignal(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})

	fake := func(ctx context.Context, idx, players int) orchestrator.GameResult {
		<-started // block until we signal

		return orchestrator.GameResult{GameIndex: idx}
	}

	pool := orchestrator.NewPool(
		orchestrator.PoolConfig{
			TargetGames:  2,
			NumPlayers:   4,
			StaggerDelay: 5 * time.Millisecond,
		},
		fake,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go pool.Run(ctx)

	// Ready should close after initial fill.
	select {
	case <-pool.Ready():
		// Good — initial 2 games launched.
	case <-time.After(2 * time.Second):
		t.Fatal("Ready() did not close in time")
	}

	cancel()
	close(started)
	pool.WaitDrain()
}

func TestPool_GracefulShutdown(t *testing.T) {
	t.Parallel()

	fake := func(ctx context.Context, idx, players int) orchestrator.GameResult {
		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
		}

		return orchestrator.GameResult{GameIndex: idx}
	}

	pool := orchestrator.NewPool(
		orchestrator.PoolConfig{
			TargetGames:  2,
			NumPlayers:   4,
			StaggerDelay: 5 * time.Millisecond,
		},
		fake,
	)

	ctx, cancel := context.WithCancel(context.Background())

	go pool.Run(ctx)
	<-pool.Ready()

	// Cancel and verify WaitDrain returns.
	cancel()

	done := make(chan struct{})
	go func() {
		pool.WaitDrain()
		close(done)
	}()

	select {
	case <-done:
		// Good.
	case <-time.After(3 * time.Second):
		t.Fatal("WaitDrain did not return after cancel")
	}
}

func TestPool_ReplacementOnCompletion(t *testing.T) {
	t.Parallel()

	var launched atomic.Int64

	fake := func(ctx context.Context, idx, players int) orchestrator.GameResult {
		launched.Add(1)
		time.Sleep(10 * time.Millisecond) // very fast games

		return orchestrator.GameResult{GameIndex: idx}
	}

	pool := orchestrator.NewPool(
		orchestrator.PoolConfig{
			TargetGames:  2,
			NumPlayers:   4,
			StaggerDelay: 2 * time.Millisecond,
		},
		fake,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go pool.Run(ctx)
	<-ctx.Done()
	pool.WaitDrain()

	// With 10ms games and 2 slots over 200ms, should launch many more than 2.
	assert.Greater(t, launched.Load(), int64(4))
}

func TestPool_ZeroTarget(t *testing.T) {
	t.Parallel()

	pool := orchestrator.NewPool(
		orchestrator.PoolConfig{TargetGames: 0},
		func(ctx context.Context, idx, players int) orchestrator.GameResult {
			t.Fatal("should not launch any games")

			return orchestrator.GameResult{}
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go pool.Run(ctx)

	// Ready should close immediately.
	select {
	case <-pool.Ready():
	case <-time.After(time.Second):
		t.Fatal("Ready() did not close for zero target")
	}

	cancel()
	pool.WaitDrain()
	assert.Empty(t, pool.Results())
}

func TestPool_IndexOffset(t *testing.T) {
	t.Parallel()

	var indices []int
	mu := make(chan int, 10)

	fake := func(ctx context.Context, idx, players int) orchestrator.GameResult {
		mu <- idx
		time.Sleep(50 * time.Millisecond)

		return orchestrator.GameResult{GameIndex: idx}
	}

	pool := orchestrator.NewPool(
		orchestrator.PoolConfig{
			TargetGames:  2,
			NumPlayers:   4,
			StaggerDelay: 5 * time.Millisecond,
			IndexOffset:  100,
		},
		fake,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	go pool.Run(ctx)
	<-ctx.Done()
	pool.WaitDrain()
	close(mu)

	for idx := range mu {
		indices = append(indices, idx)
	}

	require.NotEmpty(t, indices)
	// All indices should start from 100.
	for _, idx := range indices {
		assert.GreaterOrEqual(t, idx, 100)
	}

	assert.GreaterOrEqual(t, pool.NextIndexOffset(), 102)
}
