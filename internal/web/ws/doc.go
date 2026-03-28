// Package ws provides shared WebSocket infrastructure used by both game
// and lobby connection managers.
//
// [PlayerConnections] tracks per-player WebSocket connections for a single
// game or lobby, handling broadcast, targeted writes, connection cleanup
// on closed sockets, and ActiveConnections metric bookkeeping. [Upgrader]
// wraps the nbio WebSocket upgrader with origin checking, subprotocol
// negotiation, and keepalive pings.
//
// # Layer
//
// Web — shared WebSocket connection tracking and upgrade logic.
package ws
