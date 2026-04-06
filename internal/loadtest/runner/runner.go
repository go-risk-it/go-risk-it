package runner

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/chaos"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/orchestrator"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/userpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

// Config holds all dependencies for the event-driven runner.
type Config struct {
	BaseURL       string
	WSURL         string
	AnonKey       string
	Strategy      player.Strategy
	Timeout       time.Duration
	Accumulator   *metrics.StepAccumulator
	LiveMetrics   *metrics.LiveMetrics
	ThinkTime     time.Duration
	Timeouts      Timeouts
	ChaosInjector *chaos.Injector
	Observer      orchestrator.GameObserver
	UserPool      *userpool.Pool // nil = legacy per-game signup
}

// Runner wires all event handlers and exposes Run().
type Runner struct {
	cfg Config
	// protocolFactory allows tests to override the protocol handler.
	protocolFactory func(gameCtx *GameSession) *ProtocolHandler
	// setupOverride allows tests to skip protocol and inject state directly.
	setupOverride func(gameCtx *GameSession)
}

// New creates a Runner with default wiring.
func New(cfg Config) *Runner {
	return &Runner{cfg: cfg}
}

// newTestRunner creates a Runner that skips protocol setup and injects state.
func newTestRunner(cfg Config, setup func(gameCtx *GameSession)) *Runner {
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
//
//nolint:funlen // sequential game lifecycle
func (r *Runner) Run(ctx context.Context, gameIndex, numPlayers int) GameResult {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()

	start := time.Now()
	gameCtx := &GameSession{
		Ctx:         ctx,
		GameIndex:   gameIndex,
		StartTime:   start,
		Accumulator: r.cfg.Accumulator,
	}

	r.cfg.LiveMetrics.RecordGameStarted()
	defer r.cfg.LiveMetrics.RecordGameStopped()

	bus := NewBus()
	result := &GameResult{GameIndex: gameIndex}

	var captured bool

	if r.setupOverride != nil {
		// Test path: skip protocol, inject players directly.
		r.setupOverride(gameCtx)
		r.wireHandlers(bus, gameCtx, result)
	} else {
		// Production path: wire side-effect handlers FIRST so HealthHandler
		// registers the game before ProtocolHandler runs the entire game loop
		// synchronously (ProtocolHandler.handle emits StateReceived which chains
		// through the full move cycle until GameComplete + Stop).
		protocol := r.buildProtocolHandler(gameCtx)
		r.wireHandlers(bus, gameCtx, result)
		protocol.Register(bus)
	}

	// Register result capture AFTER wireHandlers so MetricsHandler and HealthHandler
	// see GameComplete before Stop() kills the bus. Handlers fire in registration order,
	// and Emit checks b.stopped between each — so this MUST be last.
	bus.On(EventGameComplete, func(b *Bus, e Event) {
		evt := e.(GameCompleteEvent) //nolint:forcetypeassert // event bus guarantees type

		if !captured {
			captured = true
			*result = evt.Result
			result.GameIndex = gameIndex
			result.Duration = time.Since(start)
		}

		b.Stop()
	})

	if r.setupOverride != nil {
		// Test path: emit initial state directly, then enter barrier loop.
		snap := gameCtx.Players[0].WS.View().Snapshot()
		bus.Emit(StateReceivedEvent{Snapshot: snap, Timestamp: time.Now()})
	} else {
		// Production path: ProtocolHandler sets up auth, game, WS connections.
		// No initial state emission — the barrier loop drives ALL moves,
		// including the first one, from WS notifications.
		bus.Emit(GameStartedEvent{GameIndex: gameIndex, NumPlayers: numPlayers})
	}

	// Build barrier after protocol setup (needs players' WS views).
	barrier := r.buildBarrier(ctx, gameCtx)
	defer barrier.Stop()

	// Release pooled users after WS cleanup (registered first = runs last in LIFO).
	defer func() {
		if r.cfg.UserPool != nil && len(gameCtx.AcquiredUsers) > 0 {
			r.cfg.UserPool.Release(gameCtx.AcquiredUsers)
		}
	}()

	// Cleanup: close all WS connections (registered second = runs first in LIFO).
	defer func() {
		for _, p := range gameCtx.Players {
			if p.WS != nil {
				if err := p.WS.Close(); err != nil {
					observe.Error(context.Background(), err, "ws close error",
						attribute.Int("gameIndex", gameIndex),
					)
				}
			}
		}
	}()

	// If the initial state emission already completed the game (e.g., game over
	// detected on first snapshot), skip the barrier loop.
	if captured {
		return *result
	}

	// Event-driven game loop: the UpdateBarrier waits for ALL players to
	// receive their WS broadcast after each move, then triggers the next
	// strategy cycle. This guarantees fresh state — no polling, no races.
	return r.gameLoop(ctx, bus, barrier, gameCtx, result, &captured, start)
}

// gameLoop runs the event-driven select loop: barrier signal → emit
// StateReceivedEvent → bus processes one move cycle → wait for next signal.
func (r *Runner) gameLoop(
	ctx context.Context,
	bus *Bus,
	barrier *gamestate.UpdateBarrier,
	gameCtx *GameSession,
	result *GameResult,
	captured *bool,
	start time.Time,
) GameResult {
	for {
		select {
		case _, ok := <-barrier.Signal():
			if !ok {
				if !*captured {
					result.FatalError = errors.New("barrier closed unexpectedly")
					result.Duration = time.Since(start)
				}

				return *result
			}

			now := time.Now()
			snap := gameCtx.Players[0].WS.View().Snapshot()
			bus.Emit(StateReceivedEvent{
				Snapshot:     snap,
				Timestamp:    now,
				WSReceivedAt: now,
			})

			if *captured {
				return *result
			}
		case <-ctx.Done():
			if !*captured {
				result.TimedOut = ctx.Err() == context.DeadlineExceeded
				result.Duration = time.Since(start)

				if !result.TimedOut {
					result.FatalError = ctx.Err()
				}
			}

			return *result
		}
	}
}

func (r *Runner) buildBarrier(
	ctx context.Context,
	gameCtx *GameSession,
) *gamestate.UpdateBarrier {
	channels := make([]<-chan struct{}, len(gameCtx.Players))
	for i, p := range gameCtx.Players {
		channels[i] = p.WS.View().Notify()
	}

	return gamestate.NewUpdateBarrier(ctx, channels)
}

func (r *Runner) wireHandlers(
	bus *Bus,
	gameCtx *GameSession,
	result *GameResult,
) {
	// TracingHandler MUST be first — it sets session.Ctx which other handlers read.
	tracingH := &TracingHandler{session: gameCtx}
	tracingH.Register(bus)

	// Side-effect handlers (observe events, no emissions).
	metricsH := &MetricsHandler{collector: r.cfg.Accumulator}
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
	}
	strategyH.Register(bus)

	executorH := &ExecutorHandler{gameCtx: gameCtx}
	executorH.Register(bus)

	errorH := &ErrorHandler{
		gameCtx:           gameCtx,
		result:            result,
		maxConsecutiveErr: r.cfg.Timeouts.MaxConsecutiveErr,
	}
	errorH.Register(bus)
}

func (r *Runner) buildProtocolHandler(gameCtx *GameSession) *ProtocolHandler {
	if r.protocolFactory != nil {
		return r.protocolFactory(gameCtx)
	}

	return &ProtocolHandler{
		baseURL:  r.cfg.BaseURL,
		wsURL:    r.cfg.WSURL,
		anonKey:  r.cfg.AnonKey,
		timeouts: r.cfg.Timeouts,
		gameCtx:  gameCtx,
		userPool: r.cfg.UserPool,
		newAuth: func(baseURL, anonKey string) AuthClient {
			return client.NewAuth(baseURL, anonKey)
		},
		newREST: func(baseURL, token string, collector *metrics.StepAccumulator) RESTClient {
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
				otelhttp.NewTransport(transport),
				collector,
				client.DefaultRetryConfig(),
			)
		},
		newWS: func(
			wsURL string, gameID int64, token string, collector *metrics.StepAccumulator,
		) (WSClient, error) {
			return client.ConnectWS(wsURL, gameID, token, collector)
		},
	}
}
