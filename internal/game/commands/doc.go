// Package commands defines cross-module command types for game operations.
//
// These types form the contract that the kernel router uses to dispatch
// commands from the lobby module to the game module without direct imports.
// Only command DTOs and the Handler interface live here — no business logic.
//
// # Layer
//
// API — cross-module command contract for game creation.
package commands
