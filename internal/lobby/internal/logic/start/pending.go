package start

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrStartAlreadyPending = errors.New("start already pending for this lobby")
	ErrStartTimedOut       = errors.New("game creation timed out")
)

// StartResult carries the outcome of an asynchronous game creation triggered by
// a lobby start. It is sent through the buffered channel returned by Register.
type StartResult struct {
	GameID int64
	Err    error
}

// PendingStarts is a synchronous bridge that lets the HTTP start handler wait
// for an asynchronous game-creation event. The lifecycle is:
//
//  1. Register(lobbyID) — creates a buffered(1) channel, returns it.
//  2. Await(ctx, lobbyID, ch, timeout) — blocks until Resolve sends a result,
//     the context is cancelled, or the timeout fires. Always cleans up.
//  3. Resolve(lobbyID, gameID, err) — sends the result. Fire-and-forget safe.
//  4. Cancel(lobbyID) — early cleanup without waiting.
type PendingStarts struct {
	mu      sync.Mutex
	pending map[int64]chan StartResult
}

func NewPendingStarts() *PendingStarts {
	return &PendingStarts{
		pending: make(map[int64]chan StartResult),
	}
}

// Register creates a pending start entry for the given lobby. It returns a
// receive-only channel that will carry exactly one StartResult. Returns
// ErrStartAlreadyPending if a pending start already exists for this lobby.
func (p *PendingStarts) Register(lobbyID int64) (<-chan StartResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.pending[lobbyID]; exists {
		return nil, ErrStartAlreadyPending
	}

	ch := make(chan StartResult, 1)
	p.pending[lobbyID] = ch

	return ch, nil
}

// Await blocks until a StartResult arrives on ch, the context is cancelled, or
// the timeout fires. It always removes the pending entry for lobbyID on return,
// guaranteeing cleanup regardless of outcome.
func (p *PendingStarts) Await(
	ctx context.Context,
	lobbyID int64,
	ch <-chan StartResult,
	timeout time.Duration,
) (int64, error) {
	defer p.remove(lobbyID)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case result := <-ch:
		if result.Err != nil {
			return 0, fmt.Errorf("game creation failed: %w", result.Err)
		}

		return result.GameID, nil
	case <-timer.C:
		return 0, ErrStartTimedOut
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// Resolve delivers a StartResult to the pending start for lobbyID. It is
// fire-and-forget safe: if the entry has already been removed (e.g. by a
// timeout or cancellation), the call is a no-op.
func (p *PendingStarts) Resolve(lobbyID int64, gameID int64, err error) {
	p.mu.Lock()
	ch, exists := p.pending[lobbyID]
	p.mu.Unlock()

	if !exists {
		return
	}

	// Channel is buffered(1), so this never blocks even if nobody is listening.
	ch <- StartResult{GameID: gameID, Err: err}
}

// Cancel removes the pending start entry for lobbyID without sending a result.
// Use this for early failure cleanup when the caller won't proceed to Await.
func (p *PendingStarts) Cancel(lobbyID int64) {
	p.remove(lobbyID)
}

func (p *PendingStarts) remove(lobbyID int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.pending, lobbyID)
}
