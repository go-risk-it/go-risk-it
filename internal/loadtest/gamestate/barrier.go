package gamestate

import "context"

// UpdateBarrier waits for all N player Views to receive at least one update,
// then signals on a single output channel. After signaling, it resets and
// waits for the next round of updates. This ensures the strategy only runs
// when ALL players have fresh state from the server's broadcast.
//
// The barrier uses parallel fan-in: it reads from all player channels
// simultaneously using reflect.Select, counting unique players that have
// updated per cycle. This avoids the deadlock where sequential reading
// consumed excess notifications from fast players while slow ones starved.
type UpdateBarrier struct {
	signal chan struct{}
	cancel context.CancelFunc
}

// NewUpdateBarrier creates a barrier over the given player notification and
// done channels, and starts the background fan-in goroutine.
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
// been updated. The channel is closed when the barrier goroutine exits.
func (b *UpdateBarrier) Signal() <-chan struct{} {
	return b.signal
}

// Stop cancels the background goroutine.
func (b *UpdateBarrier) Stop() {
	b.cancel()
}

type playerUpdate struct {
	player int
	dead   bool
}

func (b *UpdateBarrier) run(
	ctx context.Context,
	notifyChannels []<-chan struct{},
	doneChannels []<-chan struct{},
) {
	defer close(b.signal)

	n := len(notifyChannels)
	merged := b.startForwarders(ctx, n, notifyChannels, doneChannels)
	updated := make([]bool, n)

	for {
		ok := b.waitForAllUpdates(ctx, merged, updated, n)
		if !ok {
			return
		}

		select {
		case b.signal <- struct{}{}:
		case <-ctx.Done():
			return
		}
	}
}

// startForwarders spawns per-player goroutines that forward notifications to a
// merged channel. Returns the merged channel.
func (b *UpdateBarrier) startForwarders(
	ctx context.Context,
	n int,
	notifyChannels []<-chan struct{},
	doneChannels []<-chan struct{},
) <-chan playerUpdate {
	merged := make(chan playerUpdate, n*4)

	for i := range n {
		go func(idx int) {
			for {
				select {
				case <-notifyChannels[idx]:
					merged <- playerUpdate{player: idx}
				case <-doneChannels[idx]:
					merged <- playerUpdate{player: idx, dead: true}

					return
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	return merged
}

// waitForAllUpdates blocks until all N players have sent at least one update.
// Returns false if the barrier should stop (WS death or context cancellation).
func (b *UpdateBarrier) waitForAllUpdates(
	ctx context.Context,
	merged <-chan playerUpdate,
	updated []bool,
	n int,
) bool {
	for i := range updated {
		updated[i] = false
	}

	remaining := n

	for remaining > 0 {
		select {
		case u := <-merged:
			if u.dead {
				return false
			}

			if !updated[u.player] {
				updated[u.player] = true
				remaining--
			}
		case <-ctx.Done():
			return false
		}
	}

	return true
}
