package health

import (
	"sync"
	"time"
)

const (
	// fallbackMoveInterval is used when no move data is available.
	fallbackMoveInterval = 5 * time.Second
	// minCompletedForMedian is the minimum completed games before using their
	// mean duration for zombie detection. Below this, ZombieAge is used directly.
	minCompletedForMedian = 3
)

type gameState struct {
	startedAt    time.Time
	lastMoveAt   time.Time
	moveCount    int
	currentPhase string
}

// Tracker maintains per-game health state for a pool of concurrent games.
// All methods are safe for concurrent use.
type Tracker struct {
	mu         sync.RWMutex
	thresholds Thresholds
	games      map[int]*gameState
	now        func() time.Time // clock function for testing

	// Running stats for mean move interval computation.
	totalMoveInterval time.Duration
	moveIntervalCount int64

	// Completed game durations for zombie threshold.
	completedDurations []time.Duration
}

// NewTracker creates a health tracker with the given classification thresholds.
func NewTracker(thresholds Thresholds) *Tracker {
	return &Tracker{
		thresholds: thresholds,
		games:      make(map[int]*gameState),
		now:        time.Now,
	}
}

// RegisterGame starts tracking a new game.
func (t *Tracker) RegisterGame(gameIndex int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	t.games[gameIndex] = &gameState{
		startedAt:  now,
		lastMoveAt: now,
	}
}

// RecordMove updates a game's last activity timestamp and phase.
func (t *Tracker) RecordMove(gameIndex int, phase string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	gs, ok := t.games[gameIndex]
	if !ok {
		return
	}

	now := t.now()
	interval := now.Sub(gs.lastMoveAt)
	gs.lastMoveAt = now
	gs.moveCount++
	gs.currentPhase = phase

	t.totalMoveInterval += interval
	t.moveIntervalCount++
}

// RecordPhaseChange updates a game's current phase.
func (t *Tracker) RecordPhaseChange(gameIndex int, newPhase string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if gs, ok := t.games[gameIndex]; ok {
		gs.currentPhase = newPhase
	}
}

// CompleteGame removes a game from tracking and records its duration.
func (t *Tracker) CompleteGame(gameIndex int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	gs, ok := t.games[gameIndex]
	if !ok {
		return
	}

	duration := t.now().Sub(gs.startedAt)
	t.completedDurations = append(t.completedDurations, duration)
	delete(t.games, gameIndex)
}

// Snapshot classifies all active games and returns the health distribution.
func (t *Tracker) Snapshot() Distribution {
	t.mu.RLock()
	defer t.mu.RUnlock()

	now := t.now()

	meanInterval := t.meanMoveInterval()
	zombieAge := t.zombieThreshold()

	var dist Distribution
	dist.Total = len(t.games)

	for _, gs := range t.games {
		status := t.classify(gs, now, meanInterval, zombieAge)

		switch status {
		case StatusHealthy:
			dist.Healthy++
		case StatusSlow:
			dist.Slow++
		case StatusStalled:
			dist.Stalled++
		case StatusZombie:
			dist.Zombie++
		}
	}

	return dist
}

// classify determines the health status of a single game.
func (t *Tracker) classify(
	gs *gameState,
	now time.Time,
	meanInterval time.Duration,
	zombieAge time.Duration,
) Status {
	gameAge := now.Sub(gs.startedAt)
	timeSinceLastMove := now.Sub(gs.lastMoveAt)

	// Zombie check: game running far longer than typical.
	if zombieAge > 0 && gameAge > zombieAge {
		return StatusZombie
	}

	// Stalled: no move for > StalledMultiplier * mean.
	stalledThreshold := time.Duration(float64(meanInterval) * t.thresholds.StalledMultiplier)
	if timeSinceLastMove > stalledThreshold {
		return StatusStalled
	}

	// Slow: no move for > SlowMultiplier * mean.
	slowThreshold := time.Duration(float64(meanInterval) * t.thresholds.SlowMultiplier)
	if timeSinceLastMove > slowThreshold {
		return StatusSlow
	}

	return StatusHealthy
}

// meanMoveInterval returns the average time between moves, or a fallback.
func (t *Tracker) meanMoveInterval() time.Duration {
	if t.moveIntervalCount == 0 {
		return fallbackMoveInterval
	}

	return t.totalMoveInterval / time.Duration(t.moveIntervalCount)
}

// zombieThreshold returns the age at which a game is considered a zombie.
func (t *Tracker) zombieThreshold() time.Duration {
	if t.thresholds.ZombieAge > 0 {
		return t.thresholds.ZombieAge
	}

	// Use 3x mean completed game duration, with a fallback.
	if len(t.completedDurations) < minCompletedForMedian {
		return 0 // no zombie detection until enough data
	}

	var total time.Duration
	for _, d := range t.completedDurations {
		total += d
	}

	mean := total / time.Duration(len(t.completedDurations))

	return 3 * mean
}
