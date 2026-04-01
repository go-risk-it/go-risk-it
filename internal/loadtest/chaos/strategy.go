package chaos

import (
	"errors"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
)

// Strategy wraps a player.Strategy and injects delays and errors.
type Strategy struct {
	inner     player.Strategy
	config    Config
	rng       *rand.Rand
	mu        sync.Mutex
	collector *metrics.StepAccumulator
}

// WrapStrategy creates a chaos-injecting wrapper around the given strategy.
func WrapStrategy(
	inner player.Strategy,
	cfg Config,
	collector *metrics.StepAccumulator,
) *Strategy {
	return &Strategy{
		inner:     inner,
		config:    cfg,
		rng:       rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
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

		return nil, errors.New("chaos: injected strategy error")
	}

	if slowRoll < s.config.SlowMoveRate {
		s.collector.RecordChaosEvent(metrics.ChaosEventSlowMove)
		time.Sleep(s.config.SlowMoveDelay)
	}

	return s.inner.DecideMove(snap, userID)
}
