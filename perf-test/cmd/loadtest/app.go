package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/dbstats"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/mapgraph"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player/heuristic"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player/smart"
)

// App holds all constructed dependencies for the load test.
type App struct {
	cfg       *Config
	graph     *mapgraph.Graph
	strategy  player.Strategy
	annotator *annotations.Annotator
	otel      *metrics.OTelExporter // nil if not configured
	dbStats   *dbstats.Collector    // nil if not configured
}

// NewApp loads the map, builds the strategy, and initializes optional
// dependencies (OTel, DB stats).
func NewApp(cfg *Config) (*App, error) {
	graph, err := mapgraph.LoadFromFile(cfg.Server.MapFile)
	if err != nil {
		return nil, fmt.Errorf("load map: %w", err)
	}

	log.Printf("loaded map: %d regions, %d continents", len(graph.Regions), len(graph.Continents))

	strategy, err := buildStrategy(cfg.Game.Strategy, graph)
	if err != nil {
		return nil, err
	}

	log.Printf("using strategy: %s", strategy.Name())

	app := &App{
		cfg:       cfg,
		graph:     graph,
		strategy:  strategy,
		annotator: annotations.NewAnnotator(cfg.Obs.GrafanaURL),
	}

	// Initialize OTel exporter.
	if cfg.Obs.OTelEndpoint != "" {
		otel, err := metrics.NewOTelExporter(context.Background(), cfg.Obs.OTelEndpoint)
		if err != nil {
			return nil, fmt.Errorf("otel exporter: %w", err)
		}

		app.otel = otel
		log.Printf("OTel metrics export enabled -> %s", cfg.Obs.OTelEndpoint)
	}

	// Initialize DB stats collector.
	if cfg.Obs.DBDSN != "" {
		dbStats, err := dbstats.NewCollector(cfg.Obs.DBDSN)
		if err != nil {
			log.Printf("warning: db stats disabled: %v", err)
		} else {
			app.dbStats = dbStats
			log.Printf("DB stats collection enabled")
		}
	}

	return app, nil
}

// Close shuts down optional resources.
func (a *App) Close() {
	if a.otel != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.otel.Shutdown(shutdownCtx); err != nil {
			log.Printf("otel shutdown: %v", err)
		}
	}

	if a.dbStats != nil {
		a.dbStats.Close()
	}
}

// Run dispatches to the appropriate mode handler.
func (a *App) Run(ctx context.Context) error {
	switch a.cfg.Mode {
	case "batch":
		return a.runBatch(ctx)
	case "ramp":
		return a.runRamp(ctx)
	case "staircase":
		return a.runStaircase(ctx)
	case "adaptive":
		return a.runAdaptive(ctx)
	default:
		return fmt.Errorf(
			"unknown mode: %q (valid: batch, ramp, staircase, adaptive)",
			a.cfg.Mode,
		)
	}
}

func buildStrategy(
	name string,
	graph *mapgraph.Graph,
) (player.Strategy, error) {
	switch name {
	case "heuristic":
		return heuristic.New(graph), nil
	case "beginner":
		return smart.New(graph, smart.Beginner()), nil
	case "normal":
		return smart.New(graph, smart.Normal()), nil
	case "expert":
		return smart.New(graph, smart.Expert()), nil
	default:
		return nil, fmt.Errorf(
			"unknown strategy: %q (valid: heuristic, beginner, normal, expert)",
			name,
		)
	}
}
