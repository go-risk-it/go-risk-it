package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/chaos"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/mapgraph"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player/heuristic"
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

	// Build config.
	cfg := orchestrator.Config{
		NumGames:    *games,
		NumPlayers:  *players,
		RampUp:      *ramp,
		GameTimeout: *gameTimeout,
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

		cfg = s.Config

		// Preset chaos config is used unless CLI flags override.
		if !chaosCfg.Enabled() && s.ChaosConfig.Enabled() {
			chaosCfg = s.ChaosConfig
		}

		log.Printf("using preset %q: %d games, %d players, timeout=%v, ramp=%v",
			*preset, cfg.NumGames, cfg.NumPlayers, cfg.GameTimeout, cfg.RampUp)
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

	// Create shared components.
	collector := metrics.NewCollector(cfg.GameTimeout)

	var strategy player.Strategy = heuristic.New(graph)
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
	results := orchestrator.Run(cfg, runner, collector)
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

	// Exit with error if any game had a fatal error.
	if fatalErrors > 0 {
		fmt.Fprintf(os.Stderr, "\n%d game(s) had fatal errors\n", fatalErrors)
		os.Exit(1)
	}
}
