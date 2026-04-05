// Package orchestration coordinates the execution of game moves through a
// transactional pipeline: validate, perform, log, check mission, walk phase
// graph, and advance. Each move runs inside a RepeatableRead transaction,
// and post-commit events (MoveCompleted) are emitted through the event bus
// for reactive consumers.
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
// [LoggingService] persists move log entries as JSON and provides retrieval
// of recent move history.
//
// [ValidationService] performs generic pre-move checks: verifying the game is
// not over, the player is in the game, and it is the player's turn.
//
// # Dependencies
//
//   - [service.Service] for move-specific perform, walk, and advance logic
//   - [state.Service] for reading current game state within the transaction
//   - [mission.Service] for checking win conditions after each move
//   - [eventbus.Publisher] for emitting post-commit domain events
//   - [metrics.Metrics] for recording move counts, durations, and game lifecycle
package orchestration
