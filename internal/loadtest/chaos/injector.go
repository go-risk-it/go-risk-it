package chaos

import (
	"context"
	"math/rand/v2"
	"sync"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"go.opentelemetry.io/otel/attribute"
)

// Injector simulates player disconnects during load tests.
type Injector struct {
	config    Config
	rng       *rand.Rand
	mu        sync.Mutex
	collector *metrics.StepAccumulator
}

// NewInjector creates a new fault injector.
func NewInjector(cfg Config, collector *metrics.StepAccumulator) *Injector {
	return &Injector{
		config:    cfg,
		rng:       rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
		collector: collector,
	}
}

// PlayerHandle is the minimal interface needed to disconnect a player.
type PlayerHandle struct {
	Name string
	WS   *client.WS
}

// MaybeDisconnect rolls the disconnect probability and, if triggered,
// disconnects a random non-active player. Returns true if a disconnect
// was injected.
func (inj *Injector) MaybeDisconnect(
	players []PlayerHandle,
	activeIdx int,
) bool {
	if inj.config.DisconnectRate <= 0 || len(players) < 2 {
		return false
	}

	inj.mu.Lock()
	roll := inj.rng.Float64()
	inj.mu.Unlock()

	if roll >= inj.config.DisconnectRate {
		return false
	}

	// Pick a random non-active player.
	candidates := make([]int, 0, len(players)-1)

	for i := range players {
		if i != activeIdx {
			candidates = append(candidates, i)
		}
	}

	if len(candidates) == 0 {
		return false
	}

	inj.mu.Lock()
	targetIdx := candidates[inj.rng.IntN(len(candidates))]
	inj.mu.Unlock()

	target := players[targetIdx]
	observe.Info(context.Background(), "chaos disconnecting player",
		attribute.String("player", target.Name),
		attribute.Int("idx", targetIdx),
	)

	if inj.config.ReconnectDelay > 0 {
		// Disrupt triggers auto-reconnect via readLoop.
		target.WS.Disrupt()
	} else {
		// Permanent disconnect.
		target.WS.Close()
	}

	inj.collector.RecordChaosEvent(metrics.ChaosEventDisconnect)

	return true
}
