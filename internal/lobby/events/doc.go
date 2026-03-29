// Package lobby defines lobby-scoped domain events and typed subscription helpers.
//
// [LobbyEvent] extends [eventbus.Event] with a LobbyID accessor. Concrete
// event types — [LobbyStateChanged] and [LobbyPlayerConnected] — are
// emitted when lobby membership changes or a player's WebSocket connects.
//
// [OnLobbyEvent] wraps [eventbus.OnEvent] with a context assertion, narrowing
// the dispatched context to [LobbyContext] before invoking the handler.
// This centralizes the assertion that all lobby bus consumers need.
//
// # Layer
//
// Events-domain — lobby-specific event types depending on events/.
package lobby
