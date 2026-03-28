// Package testonly exposes REST endpoints for E2E and component test
// scaffolding that must never be compiled into production builds.
//
// It provides two endpoints: POST /api/v1/reset (truncates all game and
// lobby tables) and POST /api/v1/setup-near-win (arranges a game state
// where the first player is one move away from winning). The [Controller]
// and [Service] are wired through the fx Module.
//
// # Layer
//
// Test — test-support HTTP handlers with direct database access.
//
// # Key Types
//
// [Controller] is the HTTP-facing interface providing ResetState and
// SetupNearWin operations for test scaffolding.
//
// [Service] is the data-access interface that truncates tables and
// arranges near-win game states via raw SQL.
//
// # Dependencies
//
//   - [db.DB] for direct database access (pool-level operations)
//   - [config.DatabaseConfig] for database connection configuration
//   - [route.Route] for registering test-only HTTP endpoints
package testonly
