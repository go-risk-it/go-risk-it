// Package orchestration coordinates the execution of game moves through an
// effects-first pipeline: validate, perform (pure), walk, advance (pure),
// check mission (pure), build state, build persistence effect, persist
// (write-only ReadCommitted TX), cache update, emit events.
//
// The pipeline uses per-game mutex serialization rather than database-level
// isolation. All computation is pure (no DB reads); the only DB interaction
// is the final Persist() call which opens a single ReadCommitted transaction
// for all writes.
//
// The package is generic over move input T and result R, allowing a single
// pipeline to handle all five move types (deploy, attack, conquer, reinforce,
// cards) with compile-time type safety.
//
// # Layer
//
// Logic — move orchestration pipeline and phase lifecycle management.
//
// # Key Types
//
// [Orchestrator] is the primary interface, parameterized by move type T and
// result type R. Its single method OrchestrateMove runs the full pipeline.
//
// [OrchestratorDeps] is an fx.In struct that bundles the non-generic
// dependencies shared by all five move-specific constructors.
//
// [DeployOrchestrator], [AttackOrchestrator], [ConquerOrchestrator],
// [ReinforceOrchestrator], and [CardsOrchestrator] are type aliases that
// bind Orchestrator to concrete move and result types.
//
// [ValidationService] performs generic pre-move checks: verifying the game is
// not over, the player is in the game, and it is the player's turn.
//
// [PersistenceEffect] is a pure data container grouping all database writes
// into 6 semantic categories ordered by foreign-key dependencies:
//  1. MoveLog — references game + player (base FKs)
//  2. MoveExecution — region mutations (troops, ownership, deployable, cards, bonuses)
//  3. Elimination — cascade data (transfer cards, reassign missions, delete spurious)
//  4. CardDraw — independent card ownership changes
//  5. PhaseTransition — creates new FK-referenced rows (phase, conquer_phase, deploy_phase)
//  6. GameConclusion — updates game row (winner), semantically final
//
// All groups are optional pointers; nil groups are skipped by Persist().
// This ordering ensures referential integrity: later groups can safely reference
// rows created by earlier groups. For example, PhaseTransition inserts new phase
// rows that MoveExecution can reference, and Elimination depends on region ownership
// changes from MoveExecution.
//
// # Dependencies
//
//   - [service.Service] for move-specific perform, walk, and advance logic
//   - [state.Service] for reading current game state on cache miss
//   - [mission.Service] for checking win conditions after each move
//   - [phase.Service] for inserting phase rows during Persist
//   - [eventbus.Publisher] for emitting post-commit domain events
//   - [metrics.Metrics] for recording move counts, durations, and game lifecycle
package orchestration
