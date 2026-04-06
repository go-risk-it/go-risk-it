package gamestate

import "context"

// UpdateBarrier waits for all N player Views to receive at least one update,
// then signals on a single output channel. After signaling, it resets and
// waits for the next round of updates. This ensures the strategy only runs
// when ALL players have fresh state from the server's broadcast.
//
// The barrier runs a background goroutine that drains each player's Notify()
// channel sequentially. Since the server broadcasts to all players per move,
// all notifications arrive within the same WS delivery window.
type UpdateBarrier struct {
	signal chan struct{}
	cancel context.CancelFunc
}

// NewUpdateBarrier creates a barrier over the given player notification channels
// and starts the background fan-in goroutine. Call Stop() to clean up.
func NewUpdateBarrier(ctx context.Context, players []<-chan struct{}) *UpdateBarrier {
	ctx, cancel := context.WithCancel(ctx)
	b := &UpdateBarrier{
		signal: make(chan struct{}, 1),
		cancel: cancel,
	}

	go b.run(ctx, players)

	return b
}

// Signal returns a channel that receives a value when all players have
// been updated. Read from this in the game loop's select statement.
func (b *UpdateBarrier) Signal() <-chan struct{} {
	return b.signal
}

// Stop cancels the background goroutine.
func (b *UpdateBarrier) Stop() {
	b.cancel()
}

func (b *UpdateBarrier) run(ctx context.Context, players []<-chan struct{}) {
	defer close(b.signal) // signal channel closes when goroutine exits

	for {
		// Wait for each player to receive at least one update.
		for _, ch := range players {
			select {
			case <-ch:
			case <-ctx.Done():
				return
			}
		}

		// All players updated — signal the game loop.
		select {
		case b.signal <- struct{}{}:
		case <-ctx.Done():
			return
		}
	}
}
