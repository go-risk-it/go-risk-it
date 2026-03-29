// Package ctx defines the typed context chain used throughout the server.
//
// The chain composes two core layers of request metadata:
// [TraceContext] (OTel span) -> [UserContext] (authenticated user ID).
// Domain modules extend this chain with their own context types:
// game/ctx.GameContext (game ID) and lobby/ctx.LobbyContext (lobby ID).
// Each layer embeds the previous one, so a GameContext satisfies both
// UserContext and TraceContext. The [Detachable] interface allows the event
// bus to copy domain metadata onto a fresh context without type-switching on
// concrete types. The [LogEnricher] interface allows the slog handler to
// extract structured attributes from any context type without concrete type
// switches.
//
// # Layer
//
// Kernel — foundational context types with no internal dependencies.
package ctx
