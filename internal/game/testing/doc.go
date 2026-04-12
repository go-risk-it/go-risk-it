// Package testing provides test infrastructure for the game module.
//
// # Layer
//
// Test — game-scoped test harnesses and generators.
//
// This package contains:
//   - invariant/: property-based testing with rapid (game state invariants)
//
// Do NOT add:
//   - Production code (this is test-only infrastructure)
//   - Mocks (use game/testmocks/ for module-scoped mocks)
//   - Unit tests for individual packages (those live alongside their code)
package testing
