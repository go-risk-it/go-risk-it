package chaos

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
)

// Strategy wraps a player.Strategy and injects delays and errors.
type Strategy struct {
	inner     player.Strategy
	config    Config
	rng       *rand.Rand
	mu        sync.Mutex
	collector *metrics.Collector
}

// WrapStrategy creates a chaos-injecting wrapper around the given strategy.
func WrapStrategy(inner player.Strategy, cfg Config, collector *metrics.Collector) *Strategy {
	return &Strategy{
		inner:     inner,
		config:    cfg,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
		collector: collector,
	}
}

func (s *Strategy) Name() string {
	return "chaos(" + s.inner.Name() + ")"
}

func (s *Strategy) DecideMove(
	snap gamestate.ViewSnapshot,
	userID string,
) (*player.Action, error) {
	s.mu.Lock()
	errorRoll := s.rng.Float64()
	slowRoll := s.rng.Float64()
	s.mu.Unlock()

	if errorRoll < s.config.ErrorMoveRate {
		s.collector.RecordChaosEvent(metrics.ChaosEventErrorMove)

		return nil, fmt.Errorf("chaos: injected strategy error")
	}

	if slowRoll < s.config.SlowMoveRate {
		s.collector.RecordChaosEvent(metrics.ChaosEventSlowMove)
		time.Sleep(s.config.SlowMoveDelay)
	}

	return s.inner.DecideMove(snap, userID)
}
