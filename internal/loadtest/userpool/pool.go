package userpool

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
)

// Entry is a pooled user that can participate in games.
type Entry struct {
	Auth *client.AuthResult
}

// Config configures the user pool.
type Config struct {
	// MaxConcurrentGames is the maximum number of simultaneous games per user.
	// When all users are at capacity, new users are created lazily.
	MaxConcurrentGames int

	// AuthFactory creates a new authenticated user. Called under no lock —
	// safe to make HTTP calls.
	AuthFactory func(ctx context.Context) (*client.AuthResult, error)
}

type poolEntry struct {
	entry       *Entry
	activeGames int
}

// Pool manages reusable authenticated users with bounded concurrency per user.
type Pool struct {
	mu      sync.Mutex
	entries []*poolEntry
	cfg     Config
}

// New creates a user pool. Users are created lazily on first Acquire.
func New(cfg Config) *Pool {
	if cfg.MaxConcurrentGames <= 0 {
		cfg.MaxConcurrentGames = 2
	}

	return &Pool{cfg: cfg}
}

// Acquire returns n users that are not at their concurrency limit.
// Creates new users lazily when all existing users are at capacity.
func (p *Pool) Acquire(ctx context.Context, n int) ([]*Entry, error) {
	p.mu.Lock()

	result := make([]*Entry, 0, n)
	needed := n

	// Scan existing users for available capacity.
	for _, pe := range p.entries {
		if needed == 0 {
			break
		}

		if pe.activeGames < p.cfg.MaxConcurrentGames {
			pe.activeGames++
			result = append(result, pe.entry)
			needed--
		}
	}

	p.mu.Unlock()

	// Create new users for the deficit (outside lock — HTTP calls).
	for range needed {
		if ctx.Err() != nil {
			// Release any already-acquired users before returning error.
			p.Release(result)

			return nil, fmt.Errorf("userpool acquire cancelled: %w", ctx.Err())
		}

		auth, err := p.cfg.AuthFactory(ctx)
		if err != nil {
			p.Release(result)

			return nil, fmt.Errorf("userpool create user: %w", err)
		}

		entry := &Entry{Auth: auth}

		p.mu.Lock()
		p.entries = append(p.entries, &poolEntry{entry: entry, activeGames: 1})
		p.mu.Unlock()

		result = append(result, entry)
	}

	return result, nil
}

// Release returns users to the pool after a game finishes.
// Safe to call with nil or empty slice.
func (p *Pool) Release(entries []*Entry) {
	if len(entries) == 0 {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Build lookup set for O(n) release.
	releasing := make(map[*Entry]struct{}, len(entries))
	for _, e := range entries {
		releasing[e] = struct{}{}
	}

	for _, pe := range p.entries {
		if _, ok := releasing[pe.entry]; ok {
			pe.activeGames--
			delete(releasing, pe.entry)

			if len(releasing) == 0 {
				return
			}
		}
	}
}

// Stats returns current pool metrics.
type Stats struct {
	TotalUsers  int
	ActiveSlots int // sum of all activeGames counters
}

// Stats returns current pool metrics.
func (p *Pool) Stats() Stats {
	p.mu.Lock()
	defer p.mu.Unlock()

	var active int
	for _, pe := range p.entries {
		active += pe.activeGames
	}

	return Stats{
		TotalUsers:  len(p.entries),
		ActiveSlots: active,
	}
}
