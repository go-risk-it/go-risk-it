// Package handlers contains event-driven consumers for game domain events.
//
// Handlers subscribe to bus events (MoveCompleted, PlayerConnected) and perform
// side effects: broadcasting state to WS clients, detecting headlines (player
// eliminations, continent ownership changes), managing game lifecycle cleanup,
// and serving state to newly connecting players.
//
// All handlers are registered via fx.Invoke in Module and subscribe using
// gameevt.OnGameEvent[E] for typed event dispatch.
//
// # Layer
//
// Game-support — event consumers for state distribution and derived detection.
package handlers
