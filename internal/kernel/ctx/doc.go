// Package ctx defines the typed context chain used throughout the server.
//
// The chain composes four layers of request metadata:
// [TraceContext] (OTel span) -> [UserContext] (authenticated user ID) ->
// [GameContext] (game ID) and [LobbyContext] (lobby ID). Each layer embeds
// the previous one, so a GameContext satisfies both UserContext and
// TraceContext. The [Detachable] interface allows the event bus to copy
// domain metadata onto a fresh context without type-switching on concrete
// types.
//
// # Layer
//
// Kernel — foundational context types with no internal dependencies.
package ctx
