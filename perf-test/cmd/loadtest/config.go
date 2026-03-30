package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/chaos"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/scenario"
)

// Config holds all parsed CLI flags and derived values.
type Config struct {
	Server ServerConfig
	Game   GameConfig
	Chaos  chaos.Config
	Obs    ObservabilityConfig
	Report ReportConfig
	Run    RunConfig
}

// ReportConfig holds output format and result-tracking settings.
type ReportConfig struct {
	Output  OutputConfig
	Journal JournalFlags
}

// RunConfig holds mode dispatch and per-mode flags.
type RunConfig struct {
	Mode      string
	Preset    string
	Batch     BatchFlags
	Staircase StaircaseFlags
	Adaptive  AdaptiveFlags
	Ramp      RampFlags

	// Internal: populated by ApplyPreset for staircase/ramp modes.
	staircaseCfg *orchestrator.StaircaseConfig
	rampCfg      *orchestrator.RampConfig
}

// BatchFlags holds batch-mode CLI flags.
type BatchFlags struct {
	NumGames int
	RampUp   time.Duration
}

// ServerConfig holds server connection settings.
type ServerConfig struct {
	URL, AnonKey, MapFile string
	WSURL                 string // computed from URL
}

// GameConfig holds per-game settings.
type GameConfig struct {
	NumPlayers  int
	GameTimeout time.Duration
	ThinkTime   time.Duration
	Strategy    string
}

// ObservabilityConfig holds OTel and Grafana settings.
type ObservabilityConfig struct {
	OTelEndpoint string
	GrafanaURL   string
	DBDSN        string
}

// OutputConfig holds output format and baseline settings.
type OutputConfig struct {
	Format       string
	SaveBaseline bool
	CompareFile  string
	BaselineName string
}

// StaircaseFlags holds staircase-specific CLI flags.
type StaircaseFlags struct {
	Steps        string // raw comma-separated, parsed later
	HoldDuration time.Duration
	StopOnBreach bool
}

// AdaptiveFlags holds adaptive-specific CLI flags.
type AdaptiveFlags struct {
	Increase int
	MaxSteps int
	MaxGames int
}

// JournalFlags holds journal save/compare settings.
type JournalFlags struct {
	Save       bool
	Name       string
	Compare    string
	Hypothesis string
}

// RampFlags holds ramp-specific CLI flags.
type RampFlags struct {
	Rate           int
	MaxGames       int
	ErrorThreshold float64
	Multiplier     float64
}

// validStrategies lists accepted strategy names.
var validStrategies = map[string]bool{
	"heuristic": true,
	"beginner":  true,
	"normal":    true,
	"expert":    true,
}

// registerFlags registers all CLI flags on the given FlagSet and wires them
// to cfg's fields. Call fs.Parse() then cfg.resolve() after this.
func registerFlags(fs *flag.FlagSet) *Config {
	cfg := &Config{}

	// Server flags.
	fs.StringVar(
		&cfg.Server.URL,
		"url",
		"http://localhost:8000",
		"Base URL of the server (Kong gateway)",
	)
	fs.StringVar(&cfg.Server.AnonKey, "anon-key", "", "Supabase anon key")
	fs.StringVar(&cfg.Server.MapFile, "map", "map.json", "Path to map.json")

	// Game flags.
	fs.IntVar(&cfg.Game.NumPlayers, "players", 4, "Number of players per game")
	fs.DurationVar(&cfg.Game.GameTimeout, "game-timeout", 10*time.Minute, "Timeout per game")
	fs.DurationVar(&cfg.Game.ThinkTime, "think-time", 0, "Artificial delay between moves")
	fs.StringVar(
		&cfg.Game.Strategy,
		"strategy",
		"heuristic",
		"Bot strategy: heuristic, beginner, normal, or expert",
	)

	// Mode flags.
	fs.StringVar(&cfg.Run.Mode, "mode", "batch", "Run mode: batch, ramp, staircase, or adaptive")
	fs.StringVar(
		&cfg.Run.Preset,
		"preset",
		"",
		"Named scenario preset (overrides games/players/timeout/ramp)",
	)
	fs.IntVar(&cfg.Run.Batch.NumGames, "games", 1, "Number of concurrent games")
	fs.DurationVar(&cfg.Run.Batch.RampUp, "ramp", 0, "Ramp-up period for starting games")

	// Output flags.
	fs.StringVar(&cfg.Report.Output.Format, "output", "text", "Output format: text or json")
	fs.BoolVar(
		&cfg.Report.Output.SaveBaseline,
		"save-baseline",
		false,
		"Save performance baseline after test completes",
	)
	fs.StringVar(
		&cfg.Report.Output.CompareFile,
		"compare",
		"",
		"Path to previous baseline file for delta comparison",
	)
	fs.StringVar(
		&cfg.Report.Output.BaselineName,
		"baseline-name",
		"",
		"Named baseline for perf-journal (saves to perf-journal/baselines/ with sequence number)",
	)

	// Chaos flags.
	fs.Float64Var(
		&cfg.Chaos.DisconnectRate,
		"chaos-disconnect",
		0,
		"Player disconnect probability per loop iteration",
	)
	fs.Float64Var(&cfg.Chaos.SlowMoveRate, "chaos-slow-move", 0, "Slow move probability")
	fs.DurationVar(&cfg.Chaos.SlowMoveDelay, "chaos-slow-delay", 2*time.Second, "Slow move delay")
	fs.Float64Var(
		&cfg.Chaos.ErrorMoveRate,
		"chaos-error-move",
		0,
		"Strategy error injection probability",
	)
	fs.DurationVar(
		&cfg.Chaos.ReconnectDelay,
		"chaos-reconnect-delay",
		2*time.Second,
		"Delay before reconnecting after chaos disconnect",
	)

	// Observability flags.
	fs.StringVar(
		&cfg.Obs.OTelEndpoint,
		"otel-endpoint",
		"",
		"OTLP HTTP endpoint for live metrics export (e.g., localhost:4318). Empty disables export.",
	)
	fs.StringVar(
		&cfg.Obs.GrafanaURL,
		"grafana-url",
		"",
		"Grafana URL for annotations (e.g., http://localhost:3000). Empty disables annotations.",
	)
	fs.StringVar(
		&cfg.Obs.DBDSN,
		"db-dsn",
		"",
		"Postgres DSN for pg_stat_statements collection. Empty disables.",
	)

	// Staircase flags.
	fs.StringVar(
		&cfg.Run.Staircase.Steps,
		"steps",
		"",
		"Comma-separated staircase step counts (e.g., 5,10,20,40)",
	)
	fs.DurationVar(
		&cfg.Run.Staircase.HoldDuration,
		"hold-duration",
		60*time.Second,
		"Hold duration per staircase step",
	)
	fs.BoolVar(
		&cfg.Run.Staircase.StopOnBreach,
		"stop-on-breach",
		true,
		"Stop staircase on first SLO breach",
	)

	// Adaptive flags.
	fs.IntVar(
		&cfg.Run.Adaptive.Increase,
		"adaptive-increase",
		5,
		"Games to add per successful step in adaptive mode",
	)
	fs.IntVar(
		&cfg.Run.Adaptive.MaxSteps,
		"adaptive-max-steps",
		20,
		"Maximum number of steps in adaptive mode",
	)
	fs.IntVar(
		&cfg.Run.Adaptive.MaxGames,
		"adaptive-max-games",
		500,
		"Hard ceiling on concurrent games in adaptive mode",
	)

	// Journal flags.
	fs.BoolVar(&cfg.Report.Journal.Save, "save-journal", false, "Save staircase journal entry")
	fs.StringVar(&cfg.Report.Journal.Name, "journal-name", "", "Name for journal entry file")
	fs.StringVar(
		&cfg.Report.Journal.Compare,
		"compare-journal",
		"",
		"Path to previous journal entry for comparison",
	)
	fs.StringVar(
		&cfg.Report.Journal.Hypothesis,
		"hypothesis",
		"",
		"Hypothesis for this optimization run (annotates session)",
	)

	// Ramp flags.
	fs.IntVar(&cfg.Run.Ramp.Rate, "ramp-rate", 10, "Games per minute (ramp mode only)")
	fs.IntVar(&cfg.Run.Ramp.MaxGames, "max-games", 100, "Maximum total games (ramp mode only)")
	fs.Float64Var(
		&cfg.Run.Ramp.ErrorThreshold,
		"error-threshold",
		0.10,
		"Error rate threshold to stop (ramp mode only)",
	)
	fs.Float64Var(
		&cfg.Run.Ramp.Multiplier,
		"ramp-multiplier",
		0,
		"Rate multiplier per minute for exponential ramp (e.g., 2.0 = double each minute). 0 = constant.",
	)

	return cfg
}

// ParseFlags registers all CLI flags on flag.CommandLine, parses os.Args,
// and returns the resolved Config.
func ParseFlags() *Config {
	cfg := registerFlags(flag.CommandLine)
	flag.Parse()
	cfg.resolve()

	return cfg
}

// resolve computes derived fields after flag parsing.
func (c *Config) resolve() {
	// Build WS URL from HTTP URL using url.Parse for idempotent conversion.
	parsed, err := url.Parse(c.Server.URL)
	if err == nil {
		switch parsed.Scheme {
		case "https":
			parsed.Scheme = "wss"
		default:
			parsed.Scheme = "ws"
		}

		c.Server.WSURL = parsed.String()
	} else {
		// Fallback: simple string replacement.
		c.Server.WSURL = strings.Replace(c.Server.URL, "http://", "ws://", 1)
		c.Server.WSURL = strings.Replace(c.Server.WSURL, "https://", "wss://", 1)
	}

	// Read anon key from env if not set via flag.
	if c.Server.AnonKey == "" {
		c.Server.AnonKey = os.Getenv("ANON_KEY")
	}
}

// ApplyPreset loads a named preset and overrides config fields.
// Must be called after ParseFlags(). No-op if Preset is empty.
func (c *Config) ApplyPreset() error {
	if c.Run.Preset == "" {
		return nil
	}

	s, err := scenario.Get(c.Run.Preset)
	if err != nil {
		return fmt.Errorf("preset: %w", err)
	}

	if s.StaircaseConfig != nil {
		c.Run.Mode = "staircase"
		c.Run.staircaseCfg = s.StaircaseConfig
	} else if s.RampConfig != nil {
		c.Run.Mode = "ramp"
		c.Run.rampCfg = s.RampConfig
		c.Game.GameTimeout = s.RampConfig.GameTimeout
	} else {
		// Batch preset.
		c.Run.Batch.NumGames = s.Config.NumGames
		c.Game.NumPlayers = s.Config.NumPlayers
		c.Game.GameTimeout = s.Config.GameTimeout
		c.Run.Batch.RampUp = s.Config.RampUp
	}

	// Preset chaos config is used unless CLI flags override.
	if !c.Chaos.Enabled() && s.ChaosConfig.Enabled() {
		c.Chaos = s.ChaosConfig
	}

	return nil
}

// Validate checks that the config is usable.
func (c *Config) Validate() error {
	if c.Server.AnonKey == "" {
		return fmt.Errorf("--anon-key flag or ANON_KEY env var is required")
	}

	if !validStrategies[c.Game.Strategy] {
		return fmt.Errorf(
			"unknown strategy: %q (valid: heuristic, beginner, normal, expert)",
			c.Game.Strategy,
		)
	}

	if c.Run.Mode == "staircase" && c.Run.staircaseCfg == nil && c.Run.Staircase.Steps == "" {
		return fmt.Errorf("staircase mode requires --preset or --steps flag")
	}

	if c.Chaos.Enabled() {
		if err := c.Chaos.Validate(); err != nil {
			return fmt.Errorf("chaos config: %w", err)
		}
	}

	return nil
}
