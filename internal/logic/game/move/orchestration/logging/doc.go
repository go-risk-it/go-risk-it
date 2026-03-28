// Package logging persists move log entries as JSON and provides retrieval
// of recent move history. Each entry records the move input, result, and
// the acting player, enabling audit trails and WebSocket replay.
//
// # Layer
//
// Logic — move log persistence and retrieval.
package logging
