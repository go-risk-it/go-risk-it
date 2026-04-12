// Package cards defines the result types for cards move execution.
//
// These are pure DTOs that flow through the move pipeline —
// from performer to orchestrator to event emission.
//
// # Layer
//
// API — domain value types. Must not import from data/ or web/.
package cards
