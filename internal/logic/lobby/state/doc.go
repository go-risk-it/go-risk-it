// Package state provides lobby state queries: fetching the lobby record
// with its participants. Returns a [Lobby] domain type containing the
// lobby ID and participant list.
//
// # Layer
//
// Logic — lobby state queries.
//
// # Key Types
//
// [Service] is the primary interface for fetching the lobby record
// with its participants.
//
// [Lobby] is the domain type containing the lobby ID and participant
// list, returned by GetLobbyState.
//
// [Participant] holds a single participant's user ID.
//
// # Dependencies
//
//   - [db.Querier] for lobby database queries
package state
