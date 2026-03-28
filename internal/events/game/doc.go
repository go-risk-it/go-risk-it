// Package game defines game-scoped domain events and typed subscription helpers.
//
// [GameEvent] extends [events.Event] with a GameID accessor. Concrete event
// types — [MoveExecuted], [PhaseTransitioned], [GameCompleted],
// [GameCreated], and [PlayerConnected] — carry the full post-commit outcome
// of game state transitions. Each implements ToRecord for structured logging.
//
// [OnGameEvent] wraps [events.OnEvent] with a context assertion, narrowing
// the dispatched context to [ctx.GameContext] before invoking the handler.
// This centralizes the assertion that all game bus consumers need.
//
// # Layer
//
// Events-domain — game-specific event types depending on events/ and data/.
package game
