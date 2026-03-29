// Package router dispatches cross-module commands via a type-switch.
//
// The [Router] accepts command values and dispatches them to the appropriate
// module handler. Currently only [commands.CreateGame] is routed (lobby → game).
// The router imports only game/commands from domain modules — it sits in the
// kernel layer as a thin dispatch mechanism with no business logic.
//
// # Layer
//
// Kernel — cross-module command dispatch infrastructure.
package router
