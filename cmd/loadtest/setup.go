package main

import (
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/chaos"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/orchestrator"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/runner"
)

// setupChaos wraps the strategy with chaos behavior and returns
// an injector if disconnect is enabled.
func (a *App) setupChaos(collector *metrics.StepAccumulator) (player.Strategy, *chaos.Injector) {
	strategy := a.strategy
	if a.cfg.Chaos.Enabled() {
		strategy = chaos.WrapStrategy(strategy, a.cfg.Chaos, collector)
	}

	var injector *chaos.Injector
	if a.cfg.Chaos.DisconnectRate > 0 {
		injector = chaos.NewInjector(a.cfg.Chaos, collector)
	}

	return strategy, injector
}

// newRunFunc creates a RunFunc with chaos wrapping and runner configuration.
func (a *App) newRunFunc(
	collector *metrics.StepAccumulator,
	timeout time.Duration,
) orchestrator.RunFunc {
	strategy, injector := a.setupChaos(collector)

	return runner.New(runner.Config{
		BaseURL:       a.cfg.Server.URL,
		WSURL:         a.cfg.Server.WSURL,
		AnonKey:       a.cfg.Server.AnonKey,
		Strategy:      strategy,
		Timeout:       timeout,
		Accumulator:   collector,
		LiveMetrics:   a.liveMetrics,
		ThinkTime:     a.cfg.Game.ThinkTime,
		Timeouts:      runner.DefaultTimeouts(),
		ChaosInjector: injector,
	}).ToRunFunc()
}
