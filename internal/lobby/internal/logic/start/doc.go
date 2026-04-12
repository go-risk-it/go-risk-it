// Package start provides lobby start eligibility checks, player listing,
// and marking a lobby as started with a game ID reference. A lobby can be
// started only by its owner when the minimum participant count is met.
//
// # Layer
//
// Logic — lobby start validation and transition.
//
// # Key Types
//
// [Service] is the primary interface for start eligibility checks,
// player listing, and marking a lobby as started with a game ID.
//
// # Dependencies
//
//   - [db.Querier] for lobby database operations
package start
