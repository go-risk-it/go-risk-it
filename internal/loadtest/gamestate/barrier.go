package gamestate

import "context"

// UpdateBarrier waits for all N player Views to receive at least one update,
// then signals on a single output channel. After signaling, it resets and
// waits for the next round of updates. This ensures the strategy only runs
// when ALL players have fresh state from the server's broadcast.
//
// The barrier also watches for WS connection death (via done channels).
// If any player's WS connection closes permanently, the barrier stops.
type UpdateBarrier struct {
	signal chan struct{}
	cancel context.CancelFunc
}

// NewUpdateBarrier creates a barrier over the given player notification and
// done channels, and starts the background fan-in goroutine. Call Stop() to
// clean up.
func NewUpdateBarrier(
	ctx context.Context,
	notifyChannels []<-chan struct{},
	doneChannels []<-chan struct{},
) *UpdateBarrier {
	ctx, cancel := context.WithCancel(ctx)
	b := &UpdateBarrier{
		signal: make(chan struct{}, 1),
		cancel: cancel,
	}

	go b.run(ctx, notifyChannels, doneChannels)

	return b
}

// Signal returns a channel that receives a value when all players have
// been updated. Read from this in the game loop's select statement.
// The channel is closed when the barrier goroutine exits.
func (b *UpdateBarrier) Signal() <-chan struct{} {
	return b.signal
}

// Stop cancels the background goroutine.
func (b *UpdateBarrier) Stop() {
	b.cancel()
}

func (b *UpdateBarrier) run(
	ctx context.Context,
	notifyChannels []<-chan struct{},
	doneChannels []<-chan struct{},
) {
	defer close(b.signal)

	for {
		// Wait for each player to receive at least one update.
		for i, ch := range notifyChannels {
			select {
			case <-ch:
			case <-doneChannels[i]:
				// WS connection died — barrier can't proceed.
				return
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
