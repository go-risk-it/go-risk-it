package orchestrator

import "time"

// Default configuration values for staircase and adaptive modes.
const (
	// DefaultNumPlayers is the default number of players per game.
	DefaultNumPlayers = 4

	// DefaultGameTimeout is the default timeout for a single game.
	DefaultGameTimeout = 10 * time.Minute

	// DefaultStaggerDelay is the default delay between starting games within a step.
	DefaultStaggerDelay = 100 * time.Millisecond

	// DefaultCooldownSec is the default cooldown between staircase steps.
	DefaultCooldownSec = 5

	// IndexOffsetMultiplier determines the offset between game indices across steps.
	// Multiplied by targetGames to leave room for replacement games within a step.
	IndexOffsetMultiplier = 10

	// DefaultAdaptiveIncrease is the default games to add per successful step.
	DefaultAdaptiveIncrease = 5

	// DefaultAdaptiveMaxSteps is the default upper bound on adaptive steps.
	DefaultAdaptiveMaxSteps = 20

	// DefaultAdaptiveMaxGames is the default hard ceiling on concurrent games.
	DefaultAdaptiveMaxGames = 500

	// DefaultAdaptiveInitialProbe is the starting concurrency when no session ceiling exists.
	DefaultAdaptiveInitialProbe = 10
)
