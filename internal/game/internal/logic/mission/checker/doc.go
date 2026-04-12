// Package checker provides the mission-checking framework. Each mission type
// implements [MissionChecker] as a pure evaluator over cached snapshot data,
// and a [Registry] maps mission types to their checkers for dispatch during
// win-condition evaluation.
//
// Checkers have no service dependencies and perform no DB access. They receive
// a [CheckContext] containing all game regions, continent definitions, and the
// current user ID, plus the player's [snapshot.PlayerMission] with type-specific
// detail.
//
// # Layer
//
// Logic — mission win-condition checking (pure evaluation).
package checker
