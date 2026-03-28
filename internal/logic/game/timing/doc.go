// Package timing tracks game start times for duration metric calculation.
// [GameTiming] records when a game starts and computes elapsed duration
// when the game completes, using a concurrent-safe sync.Map.
//
// # Layer
//
// Logic — game duration tracking for metrics.
package timing
