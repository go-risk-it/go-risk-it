package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/chaos"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/mapgraph"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player/heuristic"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player/smart"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/scenario"
)

func main() {
	url := flag.String("url", "http://localhost:8000", "Base URL of the server (Kong gateway)")
	anonKey := flag.String("anon-key", "", "Supabase anon key")
	players := flag.Int("players", 4, "Number of players per game")
	gameTimeout := flag.Duration("game-timeout", 10*time.Minute, "Timeout per game")
	mapFile := flag.String("map", "map.json", "Path to map.json")
	games := flag.Int("games", 1, "Number of concurrent games")
	ramp := flag.Duration("ramp", 0, "Ramp-up period for starting games")
	preset := flag.String(
		"preset",
		"",
		"Named scenario preset (overrides games/players/timeout/ramp)",
	)
	output := flag.String("output", "text", "Output format: text or json")
	thinkTime := flag.Duration("think-time", 0, "Artificial delay between moves")
	strategyFlag := flag.String(
		"strategy",
		"heuristic",
		"Bot strategy: heuristic, beginner, normal, or expert",
	)
	mode := flag.String("mode", "batch", "Run mode: batch or ramp")
	rampRate := flag.Int("ramp-rate", 10, "Games per minute (ramp mode only)")
	maxGames := flag.Int("max-games", 100, "Maximum total games (ramp mode only)")
	errorThreshold := flag.Float64(
		"error-threshold",
		0.10,
		"Error rate threshold to stop (ramp mode only)",
	)
	rampMultiplier := flag.Float64(
		"ramp-multiplier",
		0,
		"Rate multiplier per minute for exponential ramp (e.g., 2.0 = double each minute). 0 = constant.",
	)

	// Chaos flags.
	chaosDisconnect := flag.Float64(
		"chaos-disconnect",
		0,
		"Player disconnect probability per loop iteration",
	)
	chaosSlowMove := flag.Float64("chaos-slow-move", 0, "Slow move probability")
	chaosSlowDelay := flag.Duration("chaos-slow-delay", 2*time.Second, "Slow move delay")
	chaosErrorMove := flag.Float64("chaos-error-move", 0, "Strategy error injection probability")
	chaosReconnectDelay := flag.Duration(
		"chaos-reconnect-delay",
		2*time.Second,
		"Delay before reconnecting after chaos disconnect",
	)
	otelEndpoint := flag.String(
		"otel-endpoint",
		"",
		"OTLP HTTP endpoint for live metrics export (e.g., localhost:4318). Empty disables export.",
	)
	grafanaURL := flag.String(
		"grafana-url",
		"",
		"Grafana URL for annotations (e.g., http://localhost:3000). Empty disables annotations.",
	)
	saveBaseline := flag.Bool(
		"save-baseline",
		false,
		"Save performance baseline after test completes",
	)
	compareFile := flag.String(
		"compare",
		"",
		"Path to previous baseline file for delta comparison",
	)

	flag.Parse()

	if *anonKey == "" {
		*anonKey = os.Getenv("ANON_KEY")
	}

	if *anonKey == "" {
		log.Fatal("--anon-key flag or ANON_KEY env var is required")
	}

	// Parse map.
	graph, err := mapgraph.LoadFromFile(*mapFile)
	if err != nil {
		log.Fatalf("load map: %v", err)
	}

	log.Printf("loaded map: %d regions, %d continents", len(graph.Regions), len(graph.Continents))

	// Build batch config (may be overridden by preset).
	cfg := orchestrator.Config{
		NumGames:    *games,
		NumPlayers:  *players,
		RampUp:      *ramp,
		GameTimeout: *gameTimeout,
	}

	// Build ramp config from CLI flags (may be overridden by preset).
	rampCfg := &orchestrator.RampConfig{
		GamesPerMinute: *rampRate,
		MaxGames:       *maxGames,
		ErrorThreshold: *errorThreshold,
		GameTimeout:    *gameTimeout,
		NumPlayers:     *players,
		Multiplier:     *rampMultiplier,
	}

	// Chaos config from CLI flags.
	chaosCfg := chaos.Config{
		DisconnectRate: *chaosDisconnect,
		SlowMoveRate:   *chaosSlowMove,
		SlowMoveDelay:  *chaosSlowDelay,
		ErrorMoveRate:  *chaosErrorMove,
		ReconnectDelay: *chaosReconnectDelay,
	}

	// Override with preset if specified.
	if *preset != "" {
		s, err := scenario.Get(*preset)
		if err != nil {
			log.Fatalf("preset: %v", err)
		}

		if s.RampConfig != nil {
			// Ramp preset — force ramp mode.
			*mode = "ramp"
			rampCfg = s.RampConfig
			cfg.GameTimeout = s.RampConfig.GameTimeout

			log.Printf("using ramp preset %q: %d games/min, max %d, threshold %.0f%%",
				*preset, rampCfg.GamesPerMinute, rampCfg.MaxGames, rampCfg.ErrorThreshold*100)
		} else {
			// Batch preset.
			cfg = s.Config

			log.Printf("using preset %q: %d games, %d players, timeout=%v, ramp=%v",
				*preset, cfg.NumGames, cfg.NumPlayers, cfg.GameTimeout, cfg.RampUp)
		}

		// Preset chaos config is used unless CLI flags override.
		if !chaosCfg.Enabled() && s.ChaosConfig.Enabled() {
			chaosCfg = s.ChaosConfig
		}
	}

	if chaosCfg.Enabled() {
		if err := chaosCfg.Validate(); err != nil {
			log.Fatalf("chaos config: %v", err)
		}

		log.Printf("chaos enabled: disconnect=%.0f%% slow_move=%.0f%% error_move=%.0f%%",
			chaosCfg.DisconnectRate*100, chaosCfg.SlowMoveRate*100, chaosCfg.ErrorMoveRate*100)
	}

	// Build WS URL from HTTP URL.
	wsURL := strings.Replace(*url, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	// Determine collector max duration for throughput buckets.
	maxDuration := cfg.GameTimeout
	if *mode == "ramp" {
		// Estimate total runtime for ramp mode.
		estimated := estimateRampDuration(rampCfg)
		maxDuration = estimated + rampCfg.GameTimeout
	}

	collector := metrics.NewCollector(maxDuration)

	// Initialize OTel exporter if endpoint is provided.
	if *otelEndpoint != "" {
		ctx := context.Background()

		otelExporter, err := metrics.NewOTelExporter(ctx, *otelEndpoint)
		if err != nil {
			log.Fatalf("otel exporter: %v", err)
		}

		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := otelExporter.Shutdown(shutdownCtx); err != nil {
				log.Printf("otel shutdown: %v", err)
			}
		}()

		collector.SetOTelExporter(otelExporter)
		log.Printf("OTel metrics export enabled → %s", *otelEndpoint)
	}

	var strategy player.Strategy

	annotator := annotations.NewAnnotator(*grafanaURL)

	switch *strategyFlag {
	case "heuristic":
		strategy = heuristic.New(graph)
	case "beginner":
		strategy = smart.New(graph, smart.Beginner())
	case "normal":
		strategy = smart.New(graph, smart.Normal())
	case "expert":
		strategy = smart.New(graph, smart.Expert())
	default:
		log.Fatalf(
			"unknown strategy: %q (valid: heuristic, beginner, normal, expert)",
			*strategyFlag,
		)
	}

	log.Printf("using strategy: %s", strategy.Name())

	if chaosCfg.Enabled() {
		strategy = chaos.WrapStrategy(strategy, chaosCfg, collector)
	}

	var injector *chaos.Injector
	if chaosCfg.DisconnectRate > 0 {
		injector = chaos.NewInjector(chaosCfg, collector)
	}

	runner := orchestrator.NewGameRunner(
		*url,
		wsURL,
		*anonKey,
		strategy,
		cfg.GameTimeout,
		collector,
		*thinkTime,
		orchestrator.DefaultTimeouts(),
		injector,
	)

	// Run games.
	start := time.Now()

	var results []orchestrator.GameResult

	switch *mode {
	case "batch":
		results = orchestrator.Run(cfg, runner, collector, annotator)
	case "ramp":
		// For ramp mode, use the ramp game timeout for the runner.
		runner = orchestrator.NewGameRunner(
			*url,
			wsURL,
			*anonKey,
			strategy,
			rampCfg.GameTimeout,
			collector,
			*thinkTime,
			orchestrator.DefaultTimeouts(),
			injector,
		)
		results = orchestrator.RunContinuousRamp(*rampCfg, runner, collector, annotator)
	default:
		log.Fatalf("unknown mode: %q (valid: batch, ramp)", *mode)
	}

	totalDuration := time.Since(start)

	// Count fatal errors and build report results.
	fatalErrors := 0

	reportResults := make([]metrics.GameResult, len(results))
	for i, r := range results {
		if r.FatalError != nil {
			fatalErrors++
		}

		reportResults[i] = metrics.GameResult{
			GameIndex:  r.GameIndex,
			Duration:   r.Duration,
			Moves:      r.Moves,
			Errors:     r.Errors,
			Winner:     r.Winner,
			TimedOut:   r.TimedOut,
			FatalError: r.FatalError,
		}
	}

	// Print report.
	snap := collector.Snapshot()

	switch *output {
	case "json":
		if err := metrics.PrintJSON(os.Stdout, snap, totalDuration, fatalErrors, reportResults); err != nil {
			log.Fatalf("json report: %v", err)
		}
	case "text":
		metrics.PrintReport(os.Stdout, snap, totalDuration, fatalErrors, reportResults)
	default:
		log.Fatalf("unknown output format: %q", *output)
	}

	// Baseline operations.
	handleBaseline(
		*saveBaseline,
		*compareFile,
		*preset,
		cfg.NumPlayers,
		cfg.NumGames,
		*mode,
		snap,
		totalDuration,
	)

	// Exit with error if any game had a fatal error.
	if fatalErrors > 0 {
		fmt.Fprintf(os.Stderr, "\n%d game(s) had fatal errors\n", fatalErrors)
		os.Exit(1)
	}
}

func estimateRampDuration(cfg *orchestrator.RampConfig) time.Duration {
	if cfg.Multiplier <= 0 {
		// Constant rate.
		return time.Duration(
			cfg.MaxGames/max(cfg.GamesPerMinute, 1),
		) * time.Minute
	}

	// Exponential: sum games launched per minute until max reached.
	rate := float64(cfg.GamesPerMinute)
	total := 0
	minutes := 0

	for total < cfg.MaxGames {
		total += int(rate)
		minutes++
		rate *= cfg.Multiplier
	}

	return time.Duration(minutes) * time.Minute
}

func handleBaseline(
	saveBaselineFlag bool,
	compareFile string,
	presetName string,
	numPlayers int,
	numGames int,
	mode string,
	snap *metrics.Snapshot,
	totalDuration time.Duration,
) {
	if !saveBaselineFlag && compareFile == "" {
		return
	}

	currentBaseline := buildCurrentBaseline(
		presetName,
		numPlayers,
		numGames,
		mode,
		snap,
		totalDuration,
	)

	if saveBaselineFlag {
		path, err := baseline.Save("baselines", currentBaseline)
		if err != nil {
			log.Printf("failed to save baseline: %v", err)
		} else {
			log.Printf("baseline saved: %s", path)
		}
	}

	if compareFile != "" {
		referenceBaseline, err := baseline.Load(compareFile)
		if err != nil {
			log.Fatalf("failed to load baseline: %v", err)
		}

		fmt.Println()
		baseline.PrintComparison(os.Stdout, referenceBaseline, currentBaseline)
	}
}

func buildCurrentBaseline(
	presetName string,
	numPlayers int,
	numGames int,
	mode string,
	snap *metrics.Snapshot,
	totalDuration time.Duration,
) baseline.Baseline {
	commitSHA := "unknown"

	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		commitSHA = strings.TrimSpace(string(out))
	}

	return baseline.Baseline{
		CommitSHA: commitSHA,
		Timestamp: time.Now(),
		TestParams: baseline.TestParams{
			Preset:  presetName,
			Players: numPlayers,
			Games:   numGames,
			Mode:    mode,
		},
		Metrics:     baseline.SnapshotToMetrics(snap, totalDuration.Seconds()),
		Environment: baseline.CaptureEnvironment(),
	}
}
