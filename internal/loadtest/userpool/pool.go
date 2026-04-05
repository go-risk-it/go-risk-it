package userpool

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"go.opentelemetry.io/otel/attribute"
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

	// MaxConcurrentSignups limits parallel auth requests to avoid overwhelming
	// the auth service. Default 5.
	MaxConcurrentSignups int

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
	mu        sync.Mutex
	entries   []*poolEntry
	cfg       Config
	signupSem chan struct{} // limits concurrent auth requests
}

// New creates a user pool. Users are created lazily on first Acquire.
func New(cfg Config) *Pool {
	if cfg.MaxConcurrentGames <= 0 {
		cfg.MaxConcurrentGames = 2
	}

	if cfg.MaxConcurrentSignups <= 0 {
		cfg.MaxConcurrentSignups = 5
	}

	return &Pool{
		cfg:       cfg,
		signupSem: make(chan struct{}, cfg.MaxConcurrentSignups),
	}
}

// Acquire returns n users that are not at their concurrency limit.
// Creates new users lazily when all existing users are at capacity.
func (p *Pool) Acquire(ctx context.Context, n int) ([]*Entry, error) {
	result := p.acquireExisting(n)
	needed := n - len(result)

	if needed == 0 {
		return result, nil
	}

	created, err := p.createUsers(ctx, needed)
	if err != nil {
		p.Release(result)

		return nil, err
	}

	return append(result, created...), nil
}

// acquireExisting scans the pool for users with available capacity.
func (p *Pool) acquireExisting(n int) []*Entry {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([]*Entry, 0, n)

	for _, pe := range p.entries {
		if len(result) == n {
			break
		}

		if pe.activeGames < p.cfg.MaxConcurrentGames {
			pe.activeGames++
			result = append(result, pe.entry)
		}
	}

	return result
}

// createUsers signs up new users via the auth factory, throttled by the signup
// semaphore. Logs progress every 50 users and on completion.
func (p *Pool) createUsers(ctx context.Context, needed int) ([]*Entry, error) {
	result := make([]*Entry, 0, needed)

	for i := range needed {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("userpool acquire cancelled: %w", ctx.Err())
		}

		select {
		case p.signupSem <- struct{}{}:
		case <-ctx.Done():
			return nil, fmt.Errorf("userpool acquire cancelled: %w", ctx.Err())
		}

		auth, err := p.cfg.AuthFactory(ctx)
		<-p.signupSem

		if err != nil {
			return nil, fmt.Errorf("userpool create user: %w", err)
		}

		entry := &Entry{Auth: auth}

		p.mu.Lock()
		p.entries = append(p.entries, &poolEntry{entry: entry, activeGames: 1})
		totalUsers := len(p.entries)
		p.mu.Unlock()

		if (i+1)%50 == 0 || i == needed-1 {
			observe.Info(ctx, "userpool: creating users",
				attribute.Int("created", i+1),
				attribute.Int("needed", needed),
				attribute.Int("totalUsers", totalUsers),
			)
		}

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
