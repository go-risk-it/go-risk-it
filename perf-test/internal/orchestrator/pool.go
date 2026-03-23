package orchestrator

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// RunFunc matches GameRunner.Run signature for dependency injection.
type RunFunc func(ctx context.Context, gameIndex, numPlayers int) GameResult

// PoolConfig configures the game pool.
type PoolConfig struct {
	TargetGames  int
	NumPlayers   int
	StaggerDelay time.Duration // delay between initial game launches (default 100ms)
	IndexOffset  int           // starting game index (for cross-step uniqueness)
}

// Pool maintains exactly TargetGames concurrent games via a semaphore.
// When a game finishes, a replacement is launched immediately.
type Pool struct {
	cfg     PoolConfig
	runFunc RunFunc

	mu      sync.Mutex
	results []GameResult

	wg    sync.WaitGroup
	ready chan struct{}
	done  chan struct{} // closed when Run returns (after draining all goroutines)
	once  sync.Once

	nextIndex atomic.Int64
}

// NewPool creates a new game pool.
func NewPool(cfg PoolConfig, runFunc RunFunc) *Pool {
	if cfg.StaggerDelay == 0 {
		cfg.StaggerDelay = 100 * time.Millisecond
	}

	p := &Pool{
		cfg:     cfg,
		runFunc: runFunc,
		ready:   make(chan struct{}),
		done:    make(chan struct{}),
	}
	p.nextIndex.Store(int64(cfg.IndexOffset))

	return p
}

// Run blocks, maintaining target concurrency until ctx is cancelled.
// Run drains all goroutines before returning.
func (p *Pool) Run(ctx context.Context) {
	defer close(p.done)
	defer p.wg.Wait()

	if p.cfg.TargetGames <= 0 {
		p.once.Do(func() { close(p.ready) })

		return
	}

	sem := make(chan struct{}, p.cfg.TargetGames)

	// Initial fill with stagger.
	for i := range p.cfg.TargetGames {
		select {
		case <-ctx.Done():
			p.once.Do(func() { close(p.ready) })

			return
		default:
		}

		sem <- struct{}{}
		p.launchGame(ctx, sem)

		if i < p.cfg.TargetGames-1 {
			time.Sleep(p.cfg.StaggerDelay)
		}
	}

	p.once.Do(func() { close(p.ready) })

	// Replacement loop: keep launching until cancelled.
	for {
		select {
		case <-ctx.Done():
			return
		case sem <- struct{}{}:
			p.launchGame(ctx, sem)
		}
	}
}

// launchGame starts a game goroutine that releases its semaphore slot when done.
func (p *Pool) launchGame(ctx context.Context, sem chan struct{}) {
	idx := int(p.nextIndex.Add(1) - 1)

	p.wg.Add(1)

	go func() {
		defer p.wg.Done()
		defer func() { <-sem }()

		result := p.runFunc(ctx, idx, p.cfg.NumPlayers)

		p.mu.Lock()
		p.results = append(p.results, result)
		p.mu.Unlock()
	}()
}

// WaitDrain waits for Run to return, which includes draining all goroutines.
func (p *Pool) WaitDrain() {
	<-p.done
}

// Ready returns a channel that closes when the initial fill is complete.
func (p *Pool) Ready() <-chan struct{} {
	return p.ready
}

// Results returns all collected GameResults.
func (p *Pool) Results() []GameResult {
	p.mu.Lock()
	defer p.mu.Unlock()

	out := make([]GameResult, len(p.results))
	copy(out, p.results)

	return out
}

// GamesLaunched returns the total number of games started by this pool.
func (p *Pool) GamesLaunched() int {
	return int(p.nextIndex.Load()) - p.cfg.IndexOffset
}

// NextIndexOffset returns the next available game index, useful for chaining
// pools across staircase steps.
func (p *Pool) NextIndexOffset() int {
	return int(p.nextIndex.Load())
}

// LogProgress logs current pool status.
func (p *Pool) LogProgress(stepNum int, targetGames int) {
	p.mu.Lock()
	completed := len(p.results)
	p.mu.Unlock()

	launched := p.GamesLaunched()

	log.Printf(
		"[staircase step %d/%d] launched=%d completed=%d target=%d",
		stepNum,
		targetGames,
		launched,
		completed,
		p.cfg.TargetGames,
	)
}
