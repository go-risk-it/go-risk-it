package orchestration_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/orchestration"
	"github.com/stretchr/testify/require"
)

func TestGameLocks_SameGameSerializes(t *testing.T) {
	t.Parallel()

	locks := orchestration.NewGameLocks()

	const (
		gameID     = int64(42)
		goroutines = 100
	)

	var (
		counter   int
		waitGroup sync.WaitGroup
	)

	waitGroup.Add(goroutines)

	for range goroutines {
		go func() {
			defer waitGroup.Done()

			release := locks.Lock(gameID)
			defer release()

			// Non-atomic read-modify-write: without the lock this would race.
			v := counter
			counter = v + 1
		}()
	}

	waitGroup.Wait()
	require.Equal(t, goroutines, counter)
}

func TestGameLocks_DifferentGamesParallel(t *testing.T) {
	t.Parallel()

	locks := orchestration.NewGameLocks()

	const (
		gameA = int64(1)
		gameB = int64(2)
	)

	// Both goroutines acquire their lock and hold it until the other signals.
	// If locks were shared, this would deadlock.
	var (
		aHeld = make(chan struct{})
		bHeld = make(chan struct{})
	)

	done := make(chan struct{})

	go func() {
		release := locks.Lock(gameA)
		close(aHeld) // signal: "I hold gameA's lock"

		<-bHeld // wait for gameB to also be held simultaneously
		release()
	}()

	go func() {
		release := locks.Lock(gameB)
		close(bHeld)

		<-aHeld
		release()
	}()

	go func() {
		// Both goroutines must complete without deadlock.
		<-aHeld
		<-bHeld
		close(done)
	}()

	select {
	case <-done:
		// success — both locks held simultaneously
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock: different game IDs should not contend")
	}
}

func TestGameLocks_ReleaseIdempotent(t *testing.T) {
	t.Parallel()

	locks := orchestration.NewGameLocks()

	release := locks.Lock(99)

	// First release is normal.
	release()

	// Second release must not panic.
	require.NotPanics(t, func() {
		release()
	})
}

func TestGameLocks_LockReentryBlocks(t *testing.T) {
	t.Parallel()

	locks := orchestration.NewGameLocks()

	const gameID = int64(7)

	release := locks.Lock(gameID)

	var acquired atomic.Bool

	go func() {
		secondRelease := locks.Lock(gameID)
		acquired.Store(true)
		secondRelease()
	}()

	// Give the goroutine time to attempt acquisition.
	time.Sleep(50 * time.Millisecond)
	require.False(t, acquired.Load(), "second Lock should block while first is held")

	release()

	// After release, the second goroutine should acquire promptly.
	require.Eventually(t, acquired.Load, 2*time.Second, 10*time.Millisecond)
}
