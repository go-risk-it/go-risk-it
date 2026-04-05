package orchestrator

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

// RunFunc is the signature for running a single game.
type RunFunc func(ctx context.Context, gameIndex, numPlayers int) GameResult

// PoolConfig configures the game pool.
type PoolConfig struct {
	TargetGames     int
	NumPlayers      int
	StaggerDelay    time.Duration // delay between initial game launches (default 100ms)
	FillConcurrency int           // max parallel game launches during initial fill (default 20)
	IndexOffset     int           // starting game index (for cross-step uniqueness)
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
	if cfg.FillConcurrency <= 0 {
		cfg.FillConcurrency = DefaultFillConcurrency
	}

	if cfg.StaggerDelay == 0 {
		cfg.StaggerDelay = adaptiveStagger(cfg.TargetGames, cfg.FillConcurrency)
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

	if !p.initialFill(ctx, sem) {
		return
	}

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

// initialFill launches TargetGames with bounded parallelism. Returns false
// if the context was cancelled before the fill completed.
func (p *Pool) initialFill(ctx context.Context, sem chan struct{}) bool {
	fillSem := make(chan struct{}, p.cfg.FillConcurrency)
	microStagger := p.cfg.StaggerDelay / time.Duration(p.cfg.FillConcurrency)

	for i := range p.cfg.TargetGames {
		select {
		case <-ctx.Done():
			p.once.Do(func() { close(p.ready) })

			return false
		default:
		}

		fillSem <- struct{}{}
		sem <- struct{}{}
		p.launchGameWithCallback(ctx, sem, func() { <-fillSem })

		if i < p.cfg.TargetGames-1 && microStagger > 0 {
			time.Sleep(microStagger)
		}
	}

	// Drain fillSem — all launches are in-flight or done.
	for range p.cfg.FillConcurrency {
		fillSem <- struct{}{}
	}

	p.once.Do(func() { close(p.ready) })

	return true
}

// launchGame starts a game goroutine that releases its semaphore slot when done.
func (p *Pool) launchGame(ctx context.Context, sem chan struct{}) {
	p.launchGameWithCallback(ctx, sem, nil)
}

// launchGameWithCallback starts a game goroutine. The optional onStarted callback
// is invoked once the goroutine begins executing (before runFunc). Used by the
// parallel fill to release the fill semaphore as soon as the goroutine is running.
func (p *Pool) launchGameWithCallback(
	ctx context.Context,
	sem chan struct{},
	onStarted func(),
) {
	idx := int(p.nextIndex.Add(1) - 1)

	p.wg.Go(func() {
		if onStarted != nil {
			onStarted()
		}

		defer func() { <-sem }()

		result := p.runFunc(ctx, idx, p.cfg.NumPlayers)

		p.mu.Lock()
		p.results = append(p.results, result)
		p.mu.Unlock()
	})
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

	observe.Info(context.Background(), "staircase step progress",
		attribute.Int("step", stepNum),
		attribute.Int("total_steps", targetGames),
		attribute.Int("launched", launched),
		attribute.Int("completed", completed),
		attribute.Int("target", p.cfg.TargetGames),
	)
}

// adaptiveStagger computes a stagger delay that fills the pool in roughly
// targetFillTime regardless of pool size. The stagger is clamped between
// minStagger and maxStagger to avoid either hammering the server (too low)
// or wasting the hold window (too high).
//
// Fill time ≈ targetGames × stagger / fillConcurrency.
func adaptiveStagger(targetGames, fillConcurrency int) time.Duration {
	const (
		targetFillTime = 45 * time.Second
		minStagger     = 10 * time.Millisecond
		maxStagger     = 200 * time.Millisecond
	)

	if targetGames <= 0 {
		return DefaultStaggerDelay
	}

	// stagger = targetFillTime * fillConcurrency / targetGames
	stagger := time.Duration(
		int64(targetFillTime) * int64(fillConcurrency) / int64(targetGames),
	)

	stagger = max(stagger, minStagger)
	stagger = min(stagger, maxStagger)

	return stagger
}
