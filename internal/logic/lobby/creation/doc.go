// Package creation handles lobby creation: inserting the lobby record,
// adding the owner as the first participant, and setting the lobby owner
// reference. The operation runs inside a single transaction.
//
// # Layer
//
// Logic — lobby creation orchestration.
//
// # Key Types
//
// [Service] is the primary interface providing CreateLobby and
// CreateLobbyWithQuerier for transactional lobby setup.
//
// # Dependencies
//
//   - [db.Querier] for lobby database operations
//   - [metrics.Metrics] for recording transaction metrics
package creation
