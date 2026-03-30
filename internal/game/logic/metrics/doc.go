// Package metrics provides game-specific OTel metrics ([GameMetrics]) for
// tracking active games, moves, phase durations, and game lifecycle, plus
// [GameTiming] for recording game start times and computing elapsed duration.
//
// # Layer
//
// Logic — game-domain metrics and timing with no web or data imports.
package metrics
