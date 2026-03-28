// Package pool creates and manages pgx connection pools.
//
// [NewConnectionPool] builds a pgxpool.Pool from database configuration,
// executes optional initialization SQL (e.g. SET search_path), registers an
// fx shutdown hook, and starts a background goroutine that exports pool
// statistics (active, idle, total connections, acquire latency) as
// OpenTelemetry gauge and counter metrics.
//
// # Layer
//
// Data — connection pool lifecycle and observability.
package pool
