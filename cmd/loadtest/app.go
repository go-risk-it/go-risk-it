package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/annotations"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/dbstats"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/mapgraph"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player/heuristic"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player/smart"
	"go.opentelemetry.io/otel/attribute"
)

const (
	modeBatch     = "batch"
	modeRamp      = "ramp"
	modeStaircase = "staircase"
	modeAdaptive  = "adaptive"

	strategyHeuristic = "heuristic"
	strategyBeginner  = "beginner"
	strategyNormal    = "normal"
	strategyExpert    = "expert"
)

// App holds all constructed dependencies for the load test.
type App struct {
	cfg         *Config
	graph       *mapgraph.Graph
	strategy    player.Strategy
	annotator   *annotations.Annotator
	liveMetrics *metrics.LiveMetrics // nil if OTel not configured
	dbStats     *dbstats.Collector   // nil if not configured
}

// NewApp loads the map, builds the strategy, and initializes optional
// dependencies (OTel, DB stats).
func NewApp(cfg *Config) (*App, error) {
	ctx := context.Background()

	graph, err := mapgraph.LoadFromFile(cfg.Server.MapFile)
	if err != nil {
		return nil, fmt.Errorf("load map: %w", err)
	}

	observe.Info(ctx, "loaded map",
		attribute.Int("regions", len(graph.Regions)),
		attribute.Int("continents", len(graph.Continents)),
	)

	strategy, err := buildStrategy(cfg.Game.Strategy, graph)
	if err != nil {
		return nil, err
	}

	observe.Info(ctx, "using strategy", attribute.String("name", strategy.Name()))

	app := &App{
		cfg:       cfg,
		graph:     graph,
		strategy:  strategy,
		annotator: annotations.NewAnnotator(cfg.Obs.GrafanaURL),
	}

	// Initialize OTel exporter and LiveMetrics facade.
	if cfg.Obs.OTelEndpoint != "" {
		otel, err := metrics.NewOTelExporter(ctx, cfg.Obs.OTelEndpoint)
		if err != nil {
			return nil, fmt.Errorf("otel exporter: %w", err)
		}

		app.liveMetrics = metrics.NewLiveMetrics(otel)
		observe.Info(ctx, "OTel export enabled",
			attribute.String("endpoint", cfg.Obs.OTelEndpoint),
		)
	}

	// Initialize DB stats collector.
	if cfg.Obs.DBDSN != "" {
		dbStats, err := dbstats.NewCollector(cfg.Obs.DBDSN)
		if err != nil {
			observe.Warn(ctx, "db stats disabled", attribute.String("reason", err.Error()))
		} else {
			app.dbStats = dbStats
			observe.Info(ctx, "DB stats collection enabled")
		}
	}

	return app, nil
}

// Close shuts down optional resources. Uses slog directly because OTel
// providers may already be shut down during this phase.
func (a *App) Close() {
	if a.liveMetrics != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.liveMetrics.Shutdown(shutdownCtx); err != nil {
			slog.Error("otel shutdown failed", "error", err)
		}
	}

	if a.dbStats != nil {
		if err := a.dbStats.Close(); err != nil {
			slog.Error("dbstats close failed", "error", err)
		}
	}
}

// Run dispatches to the appropriate mode handler.
func (a *App) Run(ctx context.Context) error {
	switch a.cfg.Run.Mode {
	case modeBatch:
		return a.runBatch(ctx)
	case modeRamp:
		return a.runRamp(ctx)
	case modeStaircase:
		return a.runStaircase(ctx)
	case modeAdaptive:
		return a.runAdaptive(ctx)
	default:
		return fmt.Errorf(
			"unknown mode: %q (valid: batch, ramp, staircase, adaptive)",
			a.cfg.Run.Mode,
		)
	}
}

func buildStrategy(
	name string,
	graph *mapgraph.Graph,
) (player.Strategy, error) {
	switch name {
	case strategyHeuristic:
		return heuristic.New(graph), nil
	case strategyBeginner:
		return smart.New(graph, smart.Beginner()), nil
	case strategyNormal:
		return smart.New(graph, smart.Normal()), nil
	case strategyExpert:
		return smart.New(graph, smart.Expert()), nil
	default:
		return nil, fmt.Errorf(
			"unknown strategy: %q (valid: heuristic, beginner, normal, expert)",
			name,
		)
	}
}
