package runner

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/chaos"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/client"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
)

// Config holds all dependencies for the event-driven runner.
type Config struct {
	BaseURL       string
	WSURL         string
	AnonKey       string
	Strategy      player.Strategy
	Timeout       time.Duration
	Collector     *metrics.Collector
	ThinkTime     time.Duration
	Timeouts      Timeouts
	ChaosInjector *chaos.Injector
	Observer      orchestrator.GameObserver
}

// Runner wires all event handlers and exposes Run().
type Runner struct {
	cfg Config
	// protocolFactory allows tests to override the protocol handler.
	protocolFactory func(gameCtx *GameContext) *ProtocolHandler
	// setupOverride allows tests to skip protocol and inject state directly.
	setupOverride func(gameCtx *GameContext)
}

// New creates a Runner with default wiring.
func New(cfg Config) *Runner {
	return &Runner{cfg: cfg}
}

// newTestRunner creates a Runner that skips protocol setup and injects state.
func newTestRunner(cfg Config, setup func(gameCtx *GameContext)) *Runner {
	return &Runner{cfg: cfg, setupOverride: setup}
}

// ToRunFunc returns an orchestrator.RunFunc adapter.
func (r *Runner) ToRunFunc() orchestrator.RunFunc {
	return func(ctx context.Context, gameIndex, numPlayers int) orchestrator.GameResult {
		result := r.Run(ctx, gameIndex, numPlayers)

		return orchestrator.GameResult{
			GameIndex:  result.GameIndex,
			Duration:   result.Duration,
			Moves:      result.Moves,
			Errors:     result.Errors,
			Winner:     result.Winner,
			TimedOut:   result.TimedOut,
			FatalError: result.FatalError,
		}
	}
}

// Run executes a single game with the given number of players.
func (r *Runner) Run(ctx context.Context, gameIndex, numPlayers int) GameResult {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()

	start := time.Now()
	gameCtx := &GameContext{
		GameIndex: gameIndex,
		StartTime: start,
		Collector: r.cfg.Collector,
	}

	r.cfg.Collector.RecordGameStarted()

	bus := NewBus()
	result := &GameResult{GameIndex: gameIndex}

	var captured bool

	// Register result capture — runs on GameComplete to grab the final result.
	bus.On(EventGameComplete, func(b *Bus, e Event) {
		evt := e.(GameCompleteEvent)

		if !captured {
			captured = true
			*result = evt.Result
			result.GameIndex = gameIndex
			result.Duration = time.Since(start)
		}

		b.Stop()
	})

	if r.setupOverride != nil {
		// Test path: skip protocol, inject players directly.
		r.setupOverride(gameCtx)
		r.wireHandlers(bus, gameCtx, ctx, result)

		// Emit initial state from player 0.
		snap := gameCtx.Players[0].WS.View().Snapshot()
		bus.Emit(StateReceivedEvent{Snapshot: snap, Timestamp: time.Now()})
	} else {
		// Production path: wire protocol handler + all other handlers.
		protocol := r.buildProtocolHandler(gameCtx)
		protocol.Register(bus)
		r.wireHandlers(bus, gameCtx, ctx, result)
		bus.Emit(GameStartedEvent{GameIndex: gameIndex, NumPlayers: numPlayers})
	}

	// Cleanup: close all WS connections.
	defer func() {
		for _, p := range gameCtx.Players {
			if p.WS != nil {
				if err := p.WS.Close(); err != nil {
					log.Printf("[game %d] ws close error: %v", gameIndex, err)
				}
			}
		}
	}()

	if !captured {
		result.FatalError = fmt.Errorf("no result captured (event chain stalled)")
		result.Duration = time.Since(start)
	}

	return *result
}

func (r *Runner) wireHandlers(
	bus *Bus,
	gameCtx *GameContext,
	ctx context.Context,
	result *GameResult,
) {
	// Side-effect handlers (observe events, no emissions).
	metricsH := &MetricsHandler{collector: r.cfg.Collector}
	metricsH.Register(bus)

	healthH := NewHealthHandler(r.cfg.Observer, gameCtx)
	healthH.Register(bus)

	chaosH := NewChaosHandler(r.cfg.ChaosInjector, gameCtx)
	chaosH.Register(bus)

	// Core handlers.
	strategyH := &StrategyHandler{
		strategy:  r.cfg.Strategy,
		thinkTime: r.cfg.ThinkTime,
		gameCtx:   gameCtx,
		ctx:       ctx,
	}
	strategyH.Register(bus)

	executorH := &ExecutorHandler{gameCtx: gameCtx}
	executorH.Register(bus)

	stateWatcherH := &StateWatcherHandler{
		gameCtx:  gameCtx,
		timeouts: r.cfg.Timeouts,
		ctx:      ctx,
	}
	stateWatcherH.Register(bus)

	errorH := &ErrorHandler{
		gameCtx:         gameCtx,
		timeouts:        r.cfg.Timeouts,
		result:          result,
		ctx:             ctx,
		maxStaleRetries: 5,
		maxAdvanceFails: 3,
	}
	errorH.Register(bus)
}

func (r *Runner) buildProtocolHandler(gameCtx *GameContext) *ProtocolHandler {
	if r.protocolFactory != nil {
		return r.protocolFactory(gameCtx)
	}

	return &ProtocolHandler{
		baseURL:  r.cfg.BaseURL,
		wsURL:    r.cfg.WSURL,
		anonKey:  r.cfg.AnonKey,
		timeouts: r.cfg.Timeouts,
		gameCtx:  gameCtx,
		newAuth: func(baseURL, anonKey string) AuthClient {
			return client.NewAuth(baseURL, anonKey)
		},
		newREST: func(baseURL, token string, collector *metrics.Collector) RESTClient {
			transport := &http.Transport{
				MaxIdleConnsPerHost: 4,
				IdleConnTimeout:     90 * time.Second,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
			}

			return client.NewREST(
				baseURL,
				token,
				transport,
				collector,
				client.DefaultRetryConfig(),
			)
		},
		newWS: func(
			wsURL string, gameID int64, token string, collector *metrics.Collector,
		) (WSClient, error) {
			return client.ConnectWS(wsURL, gameID, token, collector)
		},
	}
}
