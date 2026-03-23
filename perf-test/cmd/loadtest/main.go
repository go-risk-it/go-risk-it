package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/chaos"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/mapgraph"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player/heuristic"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player/smart"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/resources"
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
	baselineName := flag.String(
		"baseline-name",
		"",
		"Named baseline for perf-journal (saves to perf-journal/baselines/ with sequence number)",
	)

	// Staircase flags.
	saveJournal := flag.Bool("save-journal", false, "Save staircase journal entry")
	journalName := flag.String("journal-name", "", "Name for journal entry file")
	holdDuration := flag.Duration(
		"hold-duration",
		60*time.Second,
		"Hold duration per staircase step",
	)
	stepsFlag := flag.String(
		"steps",
		"",
		"Comma-separated staircase step counts (e.g., 5,10,20,40)",
	)
	stopOnBreach := flag.Bool("stop-on-breach", true, "Stop staircase on first SLO breach")
	compareJournal := flag.String(
		"compare-journal",
		"",
		"Path to previous journal entry for comparison",
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

	// Staircase config (may be overridden by preset).
	var staircaseCfg *orchestrator.StaircaseConfig

	// Override with preset if specified.
	if *preset != "" {
		s, err := scenario.Get(*preset)
		if err != nil {
			log.Fatalf("preset: %v", err)
		}

		if s.StaircaseConfig != nil {
			// Staircase preset — force staircase mode.
			*mode = "staircase"
			staircaseCfg = s.StaircaseConfig

			log.Printf(
				"using staircase preset %q: steps=%v, hold=%v",
				*preset, staircaseCfg.Steps, staircaseCfg.HoldDuration,
			)
		} else if s.RampConfig != nil {
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
	} else if *mode == "staircase" && staircaseCfg != nil {
		maxDuration = staircaseCfg.HoldDuration
	}

	collector := metrics.NewCollector(maxDuration)

	// Initialize OTel exporter if endpoint is provided.
	var otelExporter *metrics.OTelExporter

	if *otelEndpoint != "" {
		ctx := context.Background()

		var err error
		otelExporter, err = metrics.NewOTelExporter(ctx, *otelEndpoint)
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
		log.Printf("OTel metrics export enabled -> %s", *otelEndpoint)
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
	case "staircase":
		runStaircase(
			staircaseCfg,
			*stepsFlag,
			*holdDuration,
			*stopOnBreach,
			*url,
			wsURL,
			*anonKey,
			strategy,
			*thinkTime,
			injector,
			otelExporter,
			annotator,
			*saveJournal,
			*journalName,
			*preset,
			*compareJournal,
		)

		return
	default:
		log.Fatalf("unknown mode: %q (valid: batch, ramp, staircase)", *mode)
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

	// Run insights analysis and print results.
	metricsSnap := baseline.SnapshotToMetrics(snap, totalDuration.Seconds())
	insights := baseline.Analyze(metricsSnap)
	baseline.PrintInsights(os.Stdout, insights)

	// Baseline operations.
	handleBaseline(
		*saveBaseline,
		*compareFile,
		*baselineName,
		*preset,
		cfg.NumPlayers,
		cfg.NumGames,
		*mode,
		metricsSnap,
		insights,
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

	// Exponential: sum games launched per step until max reached.
	rate := float64(cfg.GamesPerMinute)
	total := 0
	steps := 0

	step := cfg.StepInterval
	if step <= 0 {
		step = time.Minute
	}

	for total < cfg.MaxGames {
		total += int(rate)
		steps++
		rate *= cfg.Multiplier
	}

	return time.Duration(steps) * step
}

func handleBaseline(
	saveBaselineFlag bool,
	compareFile string,
	baselineName string,
	presetName string,
	numPlayers int,
	numGames int,
	mode string,
	metricsSnap baseline.MetricsSnapshot,
	insights []baseline.Insight,
) {
	if !saveBaselineFlag && compareFile == "" {
		return
	}

	currentBaseline := buildCurrentBaseline(
		presetName,
		numPlayers,
		numGames,
		mode,
		metricsSnap,
		insights,
	)

	if saveBaselineFlag {
		var path string
		var err error

		if baselineName != "" {
			path, err = baseline.SaveNumbered(
				"perf-journal/baselines",
				baselineName,
				currentBaseline,
			)
		} else {
			path, err = baseline.Save("baselines", currentBaseline)
		}

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
	metricsSnap baseline.MetricsSnapshot,
	insights []baseline.Insight,
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
		Metrics:     metricsSnap,
		Environment: baseline.CaptureEnvironment(),
		Insights:    insights,
	}
}

//nolint:funlen,cyclop
func runStaircase(
	staircaseCfg *orchestrator.StaircaseConfig,
	stepsFlag string,
	holdDuration time.Duration,
	stopOnBreach bool,
	baseURL, wsURL, anonKey string,
	strategy player.Strategy,
	thinkTime time.Duration,
	injector *chaos.Injector,
	otelExporter *metrics.OTelExporter,
	annotator *annotations.Annotator,
	saveJournal bool,
	journalName string,
	presetName string,
	compareJournalFile string,
) {
	// Build staircase config from flags if not set by preset.
	if staircaseCfg == nil {
		steps := parseSteps(stepsFlag)
		if len(steps) == 0 {
			log.Fatal("staircase mode requires --preset or --steps flag")
		}

		staircaseCfg = &orchestrator.StaircaseConfig{
			Steps:        steps,
			HoldDuration: holdDuration,
			NumPlayers:   4,
			GameTimeout:  10 * time.Minute,
			StopOnBreach: stopOnBreach,
			StaggerDelay: 100 * time.Millisecond,
			SLOs:         baseline.DefaultSLOs(),
		}
	}

	// Set up context with signal handling (owned at the staircase level).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("[staircase] shutdown signal received, stopping...")
		cancel()
	}()

	// Build dependencies.
	deps := orchestrator.StaircaseDeps{
		RunnerFactory: func(c *metrics.Collector) orchestrator.RunFunc {
			r := orchestrator.NewGameRunner(
				baseURL, wsURL, anonKey, strategy,
				staircaseCfg.GameTimeout, c, thinkTime,
				orchestrator.DefaultTimeouts(), injector,
			)

			return r.Run
		},
		NewCollector:     metrics.NewCollector,
		CollectResources: func() resources.ServerResources { return resources.CollectServerResources(resources.DefaultStatsFunc) },
		Annotator:        annotator,
		OTelExporter:     otelExporter,
	}

	// Run staircase.
	start := time.Now()
	stepOutputs := orchestrator.RunStaircase(ctx, *staircaseCfg, deps)
	totalDuration := time.Since(start)

	// Convert StepOutputs to journal StepResults.
	slos := staircaseCfg.SLOs
	stepResults := make([]journal.StepResult, len(stepOutputs))
	levelResults := make([]baseline.LevelResult, len(stepOutputs))

	for i, so := range stepOutputs {
		metricsSnap := baseline.SnapshotToMetrics(so.Snapshot, so.Duration.Seconds())
		evalResult := slos.Evaluate(metricsSnap)

		stepResults[i] = journal.StepResult{
			TargetGames:     so.TargetGames,
			Metrics:         metricsSnap,
			SLOEval:         evalResult,
			ServerResources: so.ServerResources,
			DurationSec:     so.Duration.Seconds(),
		}

		levelResults[i] = baseline.LevelResult{
			Games:   so.TargetGames,
			Metrics: metricsSnap,
		}
	}

	// Compute SLO ceiling and breaking points.
	ceiling := journal.FindSLOCeiling(stepResults)
	breakingPoints := baseline.FindBreakingPoints(levelResults, slos)

	// Get insights from the ceiling step (or last step).
	var insights []baseline.Insight
	if len(stepResults) > 0 {
		lastIdx := len(stepResults) - 1
		if ceiling.Games > 0 {
			for i, sr := range stepResults {
				if sr.TargetGames == ceiling.Games {
					lastIdx = i

					break
				}
			}
		}

		insights = baseline.Analyze(stepResults[lastIdx].Metrics)
	}

	// Build commit SHA.
	commitSHA := "unknown"

	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		commitSHA = strings.TrimSpace(string(out))
	}

	// Build journal entry.
	entry := journal.Entry{
		CommitSHA: commitSHA,
		Timestamp: time.Now(),
		Config: journal.StaircaseParams{
			Steps:           staircaseCfg.Steps,
			HoldDurationSec: staircaseCfg.HoldDuration.Seconds(),
			NumPlayers:      staircaseCfg.NumPlayers,
			GameTimeoutSec:  staircaseCfg.GameTimeout.Seconds(),
			StopOnBreach:    staircaseCfg.StopOnBreach,
		},
		SLOCeiling:     ceiling,
		Steps:          stepResults,
		BreakingPoints: breakingPoints,
		Environment:    baseline.CaptureEnvironment(),
		Insights:       insights,
	}

	// Print report.
	journal.PrintStaircaseReport(os.Stdout, entry)

	if len(insights) > 0 {
		baseline.PrintInsights(os.Stdout, insights)
	}

	log.Printf("[staircase] complete in %v: ceiling=%d games", totalDuration, ceiling.Games)

	// Save journal entry.
	if saveJournal {
		slug := journalName
		if slug == "" {
			slug = presetName
		}

		if slug == "" {
			slug = "staircase"
		}

		path, err := journal.SaveEntry("perf-journal/entries", slug, entry)
		if err != nil {
			log.Printf("failed to save journal entry: %v", err)
		} else {
			log.Printf("journal entry saved: %s", path)
		}
	}

	// Comparison.
	if compareJournalFile != "" {
		prevEntry, err := journal.LoadEntry(compareJournalFile)
		if err != nil {
			log.Fatalf("failed to load journal entry: %v", err)
		}

		journal.PrintCeilingComparison(os.Stdout, prevEntry, entry)
	}
}

// parseSteps parses a comma-separated list of step counts.
func parseSteps(s string) []int {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	steps := make([]int, 0, len(parts))

	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			log.Fatalf("invalid step count %q: %v", p, err)
		}

		steps = append(steps, n)
	}

	return steps
}
