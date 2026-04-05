package start_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic/start"
	"github.com/stretchr/testify/require"
)

func TestRegister_Success(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()

	ch, err := ps.Register(42)

	require.NoError(t, err)
	require.NotNil(t, ch)
}

func TestRegister_Duplicate(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()

	_, err := ps.Register(42)
	require.NoError(t, err)

	ch, err := ps.Register(42)

	require.ErrorIs(t, err, start.ErrStartAlreadyPending)
	require.Nil(t, ch)
}

func TestAwait_Success(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()
	ch, err := ps.Register(42)
	require.NoError(t, err)

	go func() {
		time.Sleep(10 * time.Millisecond)
		ps.Resolve(42, 99, nil)
	}()

	gameID, err := ps.Await(context.Background(), 42, ch, 5*time.Second)

	require.NoError(t, err)
	require.Equal(t, int64(99), gameID)
}

func TestAwait_Error(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()
	ch, err := ps.Register(42)
	require.NoError(t, err)

	creationErr := errors.New("board setup failed")

	go func() {
		time.Sleep(10 * time.Millisecond)
		ps.Resolve(42, 0, creationErr)
	}()

	gameID, err := ps.Await(context.Background(), 42, ch, 5*time.Second)

	require.Error(t, err)
	require.ErrorIs(t, err, creationErr)
	require.Equal(t, int64(0), gameID)
}

func TestAwait_Timeout(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()
	ch, err := ps.Register(42)
	require.NoError(t, err)

	// Never resolve — let timeout fire.
	gameID, err := ps.Await(context.Background(), 42, ch, 10*time.Millisecond)

	require.ErrorIs(t, err, start.ErrStartTimedOut)
	require.Equal(t, int64(0), gameID)
}

func TestAwait_ContextCancelled(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()
	ch, err := ps.Register(42)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	gameID, err := ps.Await(ctx, 42, ch, 5*time.Second)

	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, int64(0), gameID)
}

func TestResolve_AfterTimeout(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()
	ch, err := ps.Register(42)
	require.NoError(t, err)

	// Let Await timeout.
	_, err = ps.Await(context.Background(), 42, ch, 10*time.Millisecond)
	require.ErrorIs(t, err, start.ErrStartTimedOut)

	// Late Resolve after entry removed — must not panic.
	require.NotPanics(t, func() {
		ps.Resolve(42, 99, nil)
	})
}

func TestResolve_UnknownLobby(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()

	// Resolve for a lobby that was never registered — must not panic.
	require.NotPanics(t, func() {
		ps.Resolve(999, 1, nil)
	})
}

func TestCancel_CleansUp(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()

	_, err := ps.Register(42)
	require.NoError(t, err)

	ps.Cancel(42)

	// Second Register should succeed after Cancel.
	ch, err := ps.Register(42)
	require.NoError(t, err)
	require.NotNil(t, ch)
}

func TestConcurrent_DifferentLobbies(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()
	const numLobbies = 100

	var wg sync.WaitGroup

	wg.Add(numLobbies)

	for i := range numLobbies {
		go func() {
			defer wg.Done()

			lobbyID := int64(i)

			ch, err := ps.Register(lobbyID)
			require.NoError(t, err)

			go func() {
				ps.Resolve(lobbyID, lobbyID*10, nil)
			}()

			gameID, err := ps.Await(context.Background(), lobbyID, ch, 5*time.Second)
			require.NoError(t, err)
			require.Equal(t, lobbyID*10, gameID)
		}()
	}

	wg.Wait()
}

func TestConcurrent_SameLobby(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()
	const lobbyID = int64(42)
	const numGoroutines = 10

	var (
		wg         sync.WaitGroup
		successes  atomic.Int32
		duplicates atomic.Int32
	)

	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()

			_, err := ps.Register(lobbyID)
			if err != nil {
				require.ErrorIs(t, err, start.ErrStartAlreadyPending)
				duplicates.Add(1)

				return
			}

			successes.Add(1)
		}()
	}

	wg.Wait()

	require.Equal(t, int32(1), successes.Load())
	require.Equal(t, int32(numGoroutines-1), duplicates.Load())
}

func TestResolve_BeforeAwait(t *testing.T) {
	t.Parallel()

	ps := start.NewPendingStarts()

	ch, err := ps.Register(42)
	require.NoError(t, err)

	// Resolve immediately, before Await is called.
	ps.Resolve(42, 99, nil)

	// Await should return instantly because the buffered channel already has a value.
	gameID, err := ps.Await(context.Background(), 42, ch, 5*time.Second)

	require.NoError(t, err)
	require.Equal(t, int64(99), gameID)
}
