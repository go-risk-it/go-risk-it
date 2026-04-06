package gamestate_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// neverDone returns N done channels that never close (no WS death in tests).
func neverDone(n int) []<-chan struct{} {
	chs := make([]<-chan struct{}, n)
	for i := range chs {
		chs[i] = make(chan struct{})
	}

	return chs
}

func TestUpdateBarrier_SignalsWhenAllPlayersUpdate(t *testing.T) {
	t.Parallel()

	ch1 := make(chan struct{}, 1)
	ch2 := make(chan struct{}, 1)

	barrier := gamestate.NewUpdateBarrier(
		context.Background(),
		[]<-chan struct{}{ch1, ch2},
		neverDone(2),
	)
	defer barrier.Stop()

	// Signal should not fire yet.
	select {
	case <-barrier.Signal():
		t.Fatal("should not signal before all players update")
	case <-time.After(10 * time.Millisecond):
	}

	// Update player 1 only.
	ch1 <- struct{}{}

	select {
	case <-barrier.Signal():
		t.Fatal("should not signal with only 1 of 2 players updated")
	case <-time.After(10 * time.Millisecond):
	}

	// Update player 2 — now both updated.
	ch2 <- struct{}{}

	select {
	case <-barrier.Signal():
		// OK
	case <-time.After(time.Second):
		t.Fatal("should signal when all players updated")
	}
}

func TestUpdateBarrier_ResetsAfterSignal(t *testing.T) {
	t.Parallel()

	ch1 := make(chan struct{}, 1)
	ch2 := make(chan struct{}, 1)

	barrier := gamestate.NewUpdateBarrier(
		context.Background(),
		[]<-chan struct{}{ch1, ch2},
		neverDone(2),
	)
	defer barrier.Stop()

	// First round.
	ch1 <- struct{}{}
	ch2 <- struct{}{}

	select {
	case <-barrier.Signal():
	case <-time.After(time.Second):
		t.Fatal("first round should signal")
	}

	// Second round — barrier should wait again.
	select {
	case <-barrier.Signal():
		t.Fatal("should not signal before second round updates")
	case <-time.After(10 * time.Millisecond):
	}

	ch1 <- struct{}{}
	ch2 <- struct{}{}

	select {
	case <-barrier.Signal():
	case <-time.After(time.Second):
		t.Fatal("second round should signal")
	}
}

func TestUpdateBarrier_StopCancelsGoroutine(t *testing.T) {
	t.Parallel()

	ch1 := make(chan struct{}, 1)

	barrier := gamestate.NewUpdateBarrier(
		context.Background(),
		[]<-chan struct{}{ch1},
		neverDone(1),
	)

	barrier.Stop()

	select {
	case _, ok := <-barrier.Signal():
		assert.False(t, ok, "signal channel should be closed after stop")
	case <-time.After(time.Second):
		t.Fatal("signal channel should close promptly after stop")
	}
}

func TestUpdateBarrier_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	ch1 := make(chan struct{}, 1)

	barrier := gamestate.NewUpdateBarrier(ctx, []<-chan struct{}{ch1}, neverDone(1))
	defer barrier.Stop()

	cancel()

	select {
	case _, ok := <-barrier.Signal():
		assert.False(t, ok, "signal channel should be closed after cancel")
	case <-time.After(time.Second):
		t.Fatal("signal channel should close promptly after cancel")
	}
}

func TestUpdateBarrier_SinglePlayer(t *testing.T) {
	t.Parallel()

	ch := make(chan struct{}, 1)

	barrier := gamestate.NewUpdateBarrier(
		context.Background(),
		[]<-chan struct{}{ch},
		neverDone(1),
	)
	defer barrier.Stop()

	ch <- struct{}{}

	select {
	case <-barrier.Signal():
	case <-time.After(time.Second):
		t.Fatal("single player barrier should signal immediately")
	}
}

func TestUpdateBarrier_FourPlayers(t *testing.T) {
	t.Parallel()

	channels := make([]chan struct{}, 4)
	readOnly := make([]<-chan struct{}, 4)

	for i := range channels {
		channels[i] = make(chan struct{}, 1)
		readOnly[i] = channels[i]
	}

	barrier := gamestate.NewUpdateBarrier(context.Background(), readOnly, neverDone(4))
	defer barrier.Stop()

	// Update 3 of 4 — should not signal.
	for i := range 3 {
		channels[i] <- struct{}{}
	}

	select {
	case <-barrier.Signal():
		t.Fatal("should not signal with 3 of 4")
	case <-time.After(10 * time.Millisecond):
	}

	// Update 4th.
	channels[3] <- struct{}{}

	select {
	case <-barrier.Signal():
	case <-time.After(time.Second):
		t.Fatal("should signal when all 4 updated")
	}

	// Verify: 3 more rounds of 4 updates.
	for range 3 {
		for _, ch := range channels {
			ch <- struct{}{}
		}

		select {
		case <-barrier.Signal():
		case <-time.After(time.Second):
			require.Fail(t, "subsequent round should signal")
		}
	}
}

func TestUpdateBarrier_WSDeath_ClosesBarrier(t *testing.T) {
	t.Parallel()

	ch1 := make(chan struct{}, 1)
	ch2 := make(chan struct{}, 1)
	done1 := make(chan struct{})
	done2 := make(chan struct{})

	barrier := gamestate.NewUpdateBarrier(
		context.Background(),
		[]<-chan struct{}{ch1, ch2},
		[]<-chan struct{}{done1, done2},
	)
	defer barrier.Stop()

	// Player 1 updates, but player 2's WS dies.
	ch1 <- struct{}{}
	close(done2)

	// Barrier should close (WS death).
	select {
	case _, ok := <-barrier.Signal():
		assert.False(t, ok, "barrier should close on WS death")
	case <-time.After(time.Second):
		t.Fatal("barrier should detect WS death promptly")
	}
}
