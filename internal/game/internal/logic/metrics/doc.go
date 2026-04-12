// Package metrics provides game-specific OTel metrics ([GameMetrics]) for
// tracking active games, game duration, per-game summary histograms (moves,
// attacks, turns, headlines), and [GameTiming] for recording game start times
// and computing elapsed duration. [GameSummaryRecorder] subscribes to bus events
// and records summary histograms at game completion.
//
// # Instruments
//
//   - game.active (UpDownCounter) — number of currently active games
//   - game.duration (Histogram) — elapsed time of completed games
//   - game.summary.moves (Histogram) — total moves per completed game
//   - game.summary.attacks (Histogram) — total attack moves per completed game
//   - game.summary.turns (Histogram) — total phase transitions per completed game
//   - game.summary.headlines (Histogram) — total headline events per completed game
//
// # Layer
//
// Logic — game-domain metrics and timing with no web or data imports.
package metrics
