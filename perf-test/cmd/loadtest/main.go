package main

import (
	"context"
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
	"github.com/go-risk-it/go-risk-it/perf-test/internal/dbstats"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal/session"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/mapgraph"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player/heuristic"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player/smart"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/resources"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/runner"
)

func main() {
	cfg := ParseFlags()
	if err := cfg.ApplyPreset(); err != nil {
		log.Fatal(err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	if cfg.Chaos.Enabled() {
		log.Printf("chaos enabled: disconnect=%.0f%% slow_move=%.0f%% error_move=%.0f%%",
			cfg.Chaos.DisconnectRate*100, cfg.Chaos.SlowMoveRate*100, cfg.Chaos.ErrorMoveRate*100)
	}

	if cfg.staircaseCfg != nil {
		log.Printf(
			"using staircase preset %q: steps=%v, hold=%v",
			cfg.Preset, cfg.staircaseCfg.Steps, cfg.staircaseCfg.HoldDuration,
		)
	} else if cfg.rampCfg != nil {
		log.Printf("using ramp preset %q: %d games/min, max %d, threshold %.0f%%",
			cfg.Preset, cfg.rampCfg.GamesPerMinute, cfg.rampCfg.MaxGames, cfg.rampCfg.ErrorThreshold*100)
	} else if cfg.Preset != "" {
		log.Printf("using preset %q: %d games, %d players, timeout=%v, ramp=%v",
			cfg.Preset, cfg.NumGames, cfg.Game.NumPlayers, cfg.Game.GameTimeout, cfg.RampUp)
	}

	// Parse map.
	graph, err := mapgraph.LoadFromFile(cfg.Server.MapFile)
	if err != nil {
		log.Fatalf("load map: %v", err)
	}

	log.Printf("loaded map: %d regions, %d continents", len(graph.Regions), len(graph.Continents))

	// Build batch config (may have been overridden by preset).
	batchCfg := orchestrator.Config{
		NumGames:    cfg.NumGames,
		NumPlayers:  cfg.Game.NumPlayers,
		RampUp:      cfg.RampUp,
		GameTimeout: cfg.Game.GameTimeout,
	}

	// Build ramp config from CLI flags (may have been overridden by preset).
	rampCfg := cfg.rampCfg
	if rampCfg == nil {
		rampCfg = &orchestrator.RampConfig{
			GamesPerMinute: cfg.Ramp.Rate,
			MaxGames:       cfg.Ramp.MaxGames,
			ErrorThreshold: cfg.Ramp.ErrorThreshold,
			GameTimeout:    cfg.Game.GameTimeout,
			NumPlayers:     cfg.Game.NumPlayers,
			Multiplier:     cfg.Ramp.Multiplier,
		}
	}

	// Determine collector max duration for throughput buckets.
	maxDuration := cfg.Game.GameTimeout
	if cfg.Mode == "ramp" {
		estimated := estimateRampDuration(rampCfg)
		maxDuration = estimated + rampCfg.GameTimeout
	} else if cfg.Mode == "staircase" && cfg.staircaseCfg != nil {
		maxDuration = cfg.staircaseCfg.HoldDuration
	} else if cfg.Mode == "adaptive" {
		maxDuration = cfg.Staircase.HoldDuration
	}

	collector := metrics.NewCollector(maxDuration)

	// Initialize OTel exporter if endpoint is provided.
	var otelExporter *metrics.OTelExporter

	if cfg.Obs.OTelEndpoint != "" {
		ctx := context.Background()

		otelExporter, err = metrics.NewOTelExporter(ctx, cfg.Obs.OTelEndpoint)
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
		log.Printf("OTel metrics export enabled -> %s", cfg.Obs.OTelEndpoint)
	}

	var strategy player.Strategy

	annotator := annotations.NewAnnotator(cfg.Obs.GrafanaURL)

	switch cfg.Game.Strategy {
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
			cfg.Game.Strategy,
		)
	}

	log.Printf("using strategy: %s", strategy.Name())

	if cfg.Chaos.Enabled() {
		strategy = chaos.WrapStrategy(strategy, cfg.Chaos, collector)
	}

	var injector *chaos.Injector
	if cfg.Chaos.DisconnectRate > 0 {
		injector = chaos.NewInjector(cfg.Chaos, collector)
	}

	// Initialize DB stats collector if DSN is provided.
	var dbStatsCollector *dbstats.Collector

	if cfg.Obs.DBDSN != "" {
		dbStatsCollector, err = dbstats.NewCollector(cfg.Obs.DBDSN)
		if err != nil {
			log.Printf("warning: db stats disabled: %v", err)
		} else {
			defer dbStatsCollector.Close()
			log.Printf("DB stats collection enabled")
		}
	}

	runGame := runner.New(runner.Config{
		BaseURL:       cfg.Server.URL,
		WSURL:         cfg.Server.WSURL,
		AnonKey:       cfg.Server.AnonKey,
		Strategy:      strategy,
		Timeout:       cfg.Game.GameTimeout,
		Collector:     collector,
		ThinkTime:     cfg.Game.ThinkTime,
		Timeouts:      runner.DefaultTimeouts(),
		ChaosInjector: injector,
	}).ToRunFunc()

	// Run games.
	start := time.Now()

	var results []orchestrator.GameResult

	switch cfg.Mode {
	case "batch":
		results = orchestrator.Run(batchCfg, runGame, collector, annotator)
	case "ramp":
		rampRunGame := runner.New(runner.Config{
			BaseURL:       cfg.Server.URL,
			WSURL:         cfg.Server.WSURL,
			AnonKey:       cfg.Server.AnonKey,
			Strategy:      strategy,
			Timeout:       rampCfg.GameTimeout,
			Collector:     collector,
			ThinkTime:     cfg.Game.ThinkTime,
			Timeouts:      runner.DefaultTimeouts(),
			ChaosInjector: injector,
		}).ToRunFunc()
		results = orchestrator.RunContinuousRamp(*rampCfg, rampRunGame, collector, annotator)
	case "staircase":
		runStaircase(cfg, strategy, injector, otelExporter, annotator, dbStatsCollector)

		return
	case "adaptive":
		runAdaptiveMode(cfg, strategy, injector, otelExporter, annotator, dbStatsCollector)

		return
	default:
		log.Fatalf("unknown mode: %q (valid: batch, ramp, staircase, adaptive)", cfg.Mode)
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

	switch cfg.Output.Format {
	case "json":
		if err := metrics.PrintJSON(os.Stdout, snap, totalDuration, fatalErrors, reportResults); err != nil {
			log.Fatalf("json report: %v", err)
		}
	case "text":
		metrics.PrintReport(os.Stdout, snap, totalDuration, fatalErrors, reportResults)
	default:
		log.Fatalf("unknown output format: %q", cfg.Output.Format)
	}

	// Run insights analysis and print results.
	metricsSnap := baseline.SnapshotToMetrics(snap, totalDuration.Seconds())
	insights := baseline.Analyze(metricsSnap)
	baseline.PrintInsights(os.Stdout, insights)

	// Baseline operations.
	handleBaseline(
		cfg.Output.SaveBaseline,
		cfg.Output.CompareFile,
		cfg.Output.BaselineName,
		cfg.Preset,
		cfg.Game.NumPlayers,
		cfg.NumGames,
		cfg.Mode,
		metricsSnap,
		insights,
	)

	// Exit with error if any game had a fatal error.
	if fatalErrors > 0 {
		fmt.Fprintf(os.Stderr, "\n%d game(s) had fatal errors\n", fatalErrors)
		os.Exit(1)
	}
}

func estimateRampDuration(rampCfg *orchestrator.RampConfig) time.Duration {
	if rampCfg.Multiplier <= 0 {
		return time.Duration(
			rampCfg.MaxGames/max(rampCfg.GamesPerMinute, 1),
		) * time.Minute
	}

	rate := float64(rampCfg.GamesPerMinute)
	total := 0
	steps := 0

	step := rampCfg.StepInterval
	if step <= 0 {
		step = time.Minute
	}

	for total < rampCfg.MaxGames {
		total += int(rate)
		steps++
		rate *= rampCfg.Multiplier
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
	cfg *Config,
	strategy player.Strategy,
	injector *chaos.Injector,
	otelExporter *metrics.OTelExporter,
	annotator *annotations.Annotator,
	dbStatsCollector *dbstats.Collector,
) {
	staircaseCfg := cfg.staircaseCfg

	// Build staircase config from flags if not set by preset.
	if staircaseCfg == nil {
		steps := parseSteps(cfg.Staircase.Steps)
		if len(steps) == 0 {
			log.Fatal("staircase mode requires --preset or --steps flag")
		}

		staircaseCfg = &orchestrator.StaircaseConfig{
			Steps:             steps,
			HoldDuration:      cfg.Staircase.HoldDuration,
			NumPlayers:        orchestrator.DefaultNumPlayers,
			GameTimeout:       orchestrator.DefaultGameTimeout,
			StopOnBreach:      cfg.Staircase.StopOnBreach,
			StaggerDelay:      orchestrator.DefaultStaggerDelay,
			SLOs:              baseline.DefaultSLOs(),
			WarmUpCompletions: cfg.Staircase.WarmupCompletions,
			WarmUpDurationSec: cfg.Staircase.WarmupDuration,
		}
	}

	// CLI warm-up flags override preset values.
	if cfg.Staircase.WarmupCompletions > 0 {
		staircaseCfg.WarmUpCompletions = cfg.Staircase.WarmupCompletions
	}

	if cfg.Staircase.WarmupDuration > 0 {
		staircaseCfg.WarmUpDurationSec = cfg.Staircase.WarmupDuration
	}

	// Set up context with signal handling.
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
		RunnerFactory: func(c *metrics.Collector, obs orchestrator.GameObserver) orchestrator.RunFunc {
			return runner.New(runner.Config{
				BaseURL:       cfg.Server.URL,
				WSURL:         cfg.Server.WSURL,
				AnonKey:       cfg.Server.AnonKey,
				Strategy:      strategy,
				Timeout:       staircaseCfg.GameTimeout,
				Collector:     c,
				ThinkTime:     cfg.Game.ThinkTime,
				Timeouts:      runner.DefaultTimeouts(),
				ChaosInjector: injector,
				Observer:      obs,
			}).ToRunFunc()
		},
		NewCollector:     metrics.NewCollector,
		CollectResources: func() resources.ServerResources { return resources.CollectServerResources(resources.DefaultStatsFunc) },
		Annotator:        annotator,
		OTelExporter:     otelExporter,
		DBStats:          dbStatsCollector,
	}

	// Run staircase.
	start := time.Now()
	stepOutputs := orchestrator.RunStaircase(ctx, *staircaseCfg, deps)
	totalDuration := time.Since(start)

	// Convert and analyze results.
	stepResults, levelResults := convertStepResults(stepOutputs, staircaseCfg.SLOs)
	ceiling := journal.FindSLOCeiling(stepResults)
	breakingPoints := baseline.FindBreakingPoints(levelResults, staircaseCfg.SLOs)
	insights := findCeilingInsights(stepResults, ceiling)
	commitSHA, branch := getGitInfo()

	// Build journal entry.
	entry := journal.Entry{
		CommitSHA: commitSHA,
		Timestamp: time.Now(),
		Branch:    branch,
		Config: journal.StaircaseParams{
			Steps:             staircaseCfg.Steps,
			HoldDurationSec:   staircaseCfg.HoldDuration.Seconds(),
			NumPlayers:        staircaseCfg.NumPlayers,
			GameTimeoutSec:    staircaseCfg.GameTimeout.Seconds(),
			StopOnBreach:      staircaseCfg.StopOnBreach,
			WarmUpCompletions: staircaseCfg.WarmUpCompletions,
			WarmUpDurationSec: staircaseCfg.WarmUpDurationSec,
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

	// Save and compare.
	if cfg.Journal.Save {
		slug := cfg.Journal.Name
		if slug == "" {
			slug = cfg.Preset
		}

		if slug == "" {
			slug = "staircase"
		}

		saveJournalEntry(entry, slug, branch, commitSHA, cfg.Journal.Hypothesis)
	}

	if cfg.Journal.Compare != "" {
		compareJournalEntries(cfg.Journal.Compare, entry)
	}
}

func handleSession(branch, commitSHA, entryPath string, ceilingGames int, hypothesis string) {
	store := session.NewStore("perf-journal/sessions")

	sess, err := store.GetOrCreate(branch, "perf-journal/entries")
	if err != nil {
		log.Printf("[session] failed to create/load session: %v", err)

		return
	}

	// Compute delta vs session baseline ceiling.
	baselineCeiling := 0
	if len(sess.Runs) > 0 {
		baselineCeiling = sess.Runs[0].CeilingGames
	}

	ref := session.RunRef{
		EntryPath:    entryPath,
		CommitSHA:    commitSHA,
		CeilingGames: ceilingGames,
		CeilingDelta: ceilingGames - baselineCeiling,
		Hypothesis:   hypothesis,
		Timestamp:    time.Now(),
	}

	if err := store.AddRun(branch, ref); err != nil {
		log.Printf("[session] failed to add run: %v", err)

		return
	}

	runNum := len(sess.Runs) + 1
	log.Printf(
		"[session] %s: run #%d, ceiling=%d (delta=%+d from session start)",
		branch, runNum, ceilingGames, ref.CeilingDelta,
	)

	if hypothesis != "" {
		log.Printf("[session] hypothesis: %s", hypothesis)
	}
}

//nolint:funlen,cyclop
func runAdaptiveMode(
	cfg *Config,
	strategy player.Strategy,
	injector *chaos.Injector,
	otelExporter *metrics.OTelExporter,
	annotator *annotations.Annotator,
	dbStatsCollector *dbstats.Collector,
) {
	// Set up context with signal handling.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("[adaptive] shutdown signal received, stopping...")
		cancel()
	}()

	// Detect branch and read session's last ceiling.
	_, branch := getGitInfo()

	initialCeiling := 0
	if branch != "" && branch != "main" && branch != "master" {
		store := session.NewStore("perf-journal/sessions")

		sess, err := store.Load(branch)
		if err == nil && sess.LastCeiling() > 0 {
			initialCeiling = sess.LastCeiling()
			log.Printf("[adaptive] using session ceiling %d as starting point", initialCeiling)
		}
	}

	// Build adaptive config.
	adaptiveCfg := orchestrator.AdaptiveConfig{
		InitialCeiling:    initialCeiling,
		AdditiveIncrease:  cfg.Adaptive.Increase,
		HoldDuration:      cfg.Staircase.HoldDuration,
		NumPlayers:        4,
		GameTimeout:       10 * time.Minute,
		StaggerDelay:      100 * time.Millisecond,
		SLOs:              baseline.DefaultSLOs(),
		MaxSteps:          cfg.Adaptive.MaxSteps,
		MaxGames:          cfg.Adaptive.MaxGames,
		WarmUpCompletions: cfg.Staircase.WarmupCompletions,
		WarmUpDurationSec: cfg.Staircase.WarmupDuration,
	}

	// Build dependencies (same as staircase).
	deps := orchestrator.StaircaseDeps{
		RunnerFactory: func(c *metrics.Collector, obs orchestrator.GameObserver) orchestrator.RunFunc {
			return runner.New(runner.Config{
				BaseURL:       cfg.Server.URL,
				WSURL:         cfg.Server.WSURL,
				AnonKey:       cfg.Server.AnonKey,
				Strategy:      strategy,
				Timeout:       adaptiveCfg.GameTimeout,
				Collector:     c,
				ThinkTime:     cfg.Game.ThinkTime,
				Timeouts:      runner.DefaultTimeouts(),
				ChaosInjector: injector,
				Observer:      obs,
			}).ToRunFunc()
		},
		NewCollector:     metrics.NewCollector,
		CollectResources: func() resources.ServerResources { return resources.CollectServerResources(resources.DefaultStatsFunc) },
		Annotator:        annotator,
		OTelExporter:     otelExporter,
		DBStats:          dbStatsCollector,
	}

	// Run adaptive staircase.
	start := time.Now()
	result := orchestrator.RunAdaptive(ctx, adaptiveCfg, deps)
	totalDuration := time.Since(start)

	// Convert and analyze results.
	stepResults, _ := convertStepResults(result.Steps, adaptiveCfg.SLOs)
	ceiling := journal.FindSLOCeiling(stepResults)

	// Override ceiling games from the adaptive result (it knows the converged value).
	if result.Ceiling > 0 && ceiling.Games == 0 {
		ceiling.Games = result.Ceiling
	}

	insights := findCeilingInsights(stepResults, ceiling)
	commitSHA, _ := getGitInfo()

	// Collect steps list for config.
	configSteps := make([]int, len(result.Steps))
	for i, so := range result.Steps {
		configSteps[i] = so.TargetGames
	}

	// Build journal entry.
	entry := journal.Entry{
		CommitSHA: commitSHA,
		Timestamp: time.Now(),
		Branch:    branch,
		Config: journal.StaircaseParams{
			Mode:              "adaptive",
			Steps:             configSteps,
			HoldDurationSec:   cfg.Staircase.HoldDuration.Seconds(),
			NumPlayers:        adaptiveCfg.NumPlayers,
			GameTimeoutSec:    adaptiveCfg.GameTimeout.Seconds(),
			WarmUpCompletions: cfg.Staircase.WarmupCompletions,
			WarmUpDurationSec: cfg.Staircase.WarmupDuration,
		},
		SLOCeiling:  ceiling,
		Steps:       stepResults,
		Environment: baseline.CaptureEnvironment(),
		Insights:    insights,
	}

	// Print report.
	journal.PrintStaircaseReport(os.Stdout, entry)

	if len(insights) > 0 {
		baseline.PrintInsights(os.Stdout, insights)
	}

	convergedStr := ""
	if result.Converged {
		convergedStr = " (converged)"
	}

	log.Printf(
		"[adaptive] complete in %v: ceiling=%d games%s",
		totalDuration,
		ceiling.Games,
		convergedStr,
	)

	// Save and compare.
	if cfg.Journal.Save {
		slug := cfg.Journal.Name
		if slug == "" {
			slug = "adaptive"
		}

		saveJournalEntry(entry, slug, branch, commitSHA, cfg.Journal.Hypothesis)
	}

	if cfg.Journal.Compare != "" {
		compareJournalEntries(cfg.Journal.Compare, entry)
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

// convertStepResults converts orchestrator StepOutputs to journal StepResults
// and baseline LevelResults.
func convertStepResults(
	outputs []orchestrator.StepOutput,
	slos baseline.SLOSet,
) ([]journal.StepResult, []baseline.LevelResult) {
	stepResults := make([]journal.StepResult, len(outputs))
	levelResults := make([]baseline.LevelResult, len(outputs))

	for i, so := range outputs {
		metricsSnap := baseline.SnapshotToMetrics(so.Snapshot, so.Duration.Seconds())
		evalResult := slos.Evaluate(metricsSnap)

		stepResults[i] = journal.StepResult{
			TargetGames:        so.TargetGames,
			Metrics:            metricsSnap,
			SLOEval:            evalResult,
			ServerResources:    so.ServerResources,
			DurationSec:        so.Duration.Seconds(),
			HealthDistribution: so.HealthDistribution,
			DBStats:            so.DBStats,
		}

		levelResults[i] = baseline.LevelResult{
			Games:   so.TargetGames,
			Metrics: metricsSnap,
		}
	}

	return stepResults, levelResults
}

// findCeilingInsights returns insights from the step that matches the ceiling,
// or the last step if no ceiling.
func findCeilingInsights(
	stepResults []journal.StepResult,
	ceiling journal.SLOCeiling,
) []baseline.Insight {
	if len(stepResults) == 0 {
		return nil
	}

	lastIdx := len(stepResults) - 1
	if ceiling.Games > 0 {
		for i, sr := range stepResults {
			if sr.TargetGames == ceiling.Games {
				lastIdx = i

				break
			}
		}
	}

	return baseline.Analyze(stepResults[lastIdx].Metrics)
}

// getGitInfo returns the current short commit SHA and branch name.
func getGitInfo() (commitSHA, branch string) {
	commitSHA = "unknown"

	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		commitSHA = strings.TrimSpace(string(out))
	}

	branchOut, err := exec.Command("git", "branch", "--show-current").Output()
	if err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}

	return commitSHA, branch
}

// saveJournalEntry saves the journal entry and handles session integration.
func saveJournalEntry(
	entry journal.Entry,
	slug string,
	branch, commitSHA, hypothesis string,
) {
	path, err := journal.SaveEntry("perf-journal/entries", slug, entry)
	if err != nil {
		log.Printf("failed to save journal entry: %v", err)

		return
	}

	log.Printf("journal entry saved: %s", path)

	if branch != "" && branch != "main" && branch != "master" {
		handleSession(branch, commitSHA, path, entry.SLOCeiling.Games, hypothesis)
	}
}

// compareJournalEntries prints a comparison between two journal entries.
func compareJournalEntries(compareFile string, entry journal.Entry) {
	prevEntry, err := journal.LoadEntry(compareFile)
	if err != nil {
		log.Fatalf("failed to load journal entry: %v", err)
	}

	journal.PrintCeilingComparison(os.Stdout, prevEntry, entry)
}
