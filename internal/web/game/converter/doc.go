// Package converter transforms game snapshots into pre-serialized WebSocket
// messages for client delivery.
//
// [ConvertPublicSnapshot] produces broadcast-ready messages (game state, board
// state, player state) from a [snapshot.PublicSnapshot]. [ConvertPrivateSnapshot]
// produces per-player messages (card state, mission state) from a
// [snapshot.PrivateSnapshot], using a [MissionResolver] callback to fetch
// mission details without importing the controller package directly.
//
// # Layer
//
// Web — snapshot-to-wire serialization for WebSocket delivery.
package converter
