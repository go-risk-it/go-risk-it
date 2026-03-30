// Package management handles lobby participation: joining a lobby as a new
// participant and querying the lobbies visible to a user (owned, joined,
// and joinable). A LobbyStateChanged event is emitted after a successful
// join for WebSocket broadcast.
//
// # Layer
//
// Logic — lobby join and lobby listing.
//
// # Key Types
//
// [Service] is the primary interface for joining lobbies and querying
// the lobbies visible to a user.
//
// [UserLobbies] groups the three lobby views a user can see: owned,
// joined, and joinable.
//
// # Dependencies
//
//   - [eventbus.Publisher] for emitting LobbyStateChanged events after a join
//   - [db.Querier] for lobby database operations
//   - [metrics.Metrics] for recording transaction metrics
package management
