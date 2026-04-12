// Package invariant implements a property-based testing framework for the Risk game engine.
//
// The framework simulates complete games with randomized moves and checks 12 game-state
// invariants after every move, catching emergent rule violations that unit tests miss.
//
// # Layer
//
// Test — property-based game invariant testing framework.
//
// # Key Types
//
// [Harness] owns the test infrastructure: a testcontainer Postgres database, schema migrations,
// and an fx-wired set of game services. A single Harness is shared across all tests via TestMain.
//
// [GameSnapshot] captures the full game state (phase, turn, regions, players, card count) at a
// point in time. All invariant checkers receive snapshots, never raw database handles.
//
// [Generator] produces valid random moves for each game phase using a seeded PCG random source.
//
// [SimulationConfig] and [SimulationResult] control and report on a single game simulation run.
// [RunGame] creates a game, plays moves up to MaxMoves, and checks [CheckAll] after every move.
//
// [RapidSimulationConfig] and [RunGameProperty] integrate with [pgregory.net/rapid] for
// property-based testing with automatic seed shrinking on failure.
//
// # Invariants
//
// The 12 invariants registered in [AllInvariants]:
//
//   - EveryRegionHasMinOneTroop: every region has at least 1 troop (CONQUER allows one at 0)
//   - EveryRegionHasExactlyOneOwner: no unowned regions
//   - RegionCountEquals42: the board always has exactly 42 regions
//   - PhaseIsValid: current phase is CARDS, DEPLOY, ATTACK, CONQUER, or REINFORCE
//   - TurnNeverDecreases: turn number is monotonically non-decreasing
//   - TroopDeltaMatchesPhase: troop changes are consistent with the preceding phase
//   - AllRegionsAccountedForInPlayerCounts: player RegionCount matches actual board state
//   - EliminatedPlayersOwnNoRegions: eliminated players own zero regions
//   - TroopConservation: total troops never fall below 42 (one per region at game start)
//   - CardDeckConservation: total cards always equal 44 (42 region + 2 joker)
//   - PhaseTransitionLegality: transitions follow the valid phase state machine
//   - TerritoryIntegrity: 42 regions, each with an owner, none with negative troops
//
// # Running
//
// These tests require Docker (for testcontainers) and the invariant build tag:
//
//	go test -tags invariant -v -timeout 300s ./internal/testing/invariant/
package invariant
