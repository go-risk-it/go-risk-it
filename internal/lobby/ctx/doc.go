// Package ctx defines the lobby-scoped context type that enriches
// [kernelctx.UserContext] with a lobby ID. [LobbyContext] composes user
// identity, OTel tracing, and lobby-scoped metadata so that every layer of
// the lobby module has typed access to the active lobby. [WithLobbyID]
// constructs a LobbyContext from an authenticated user context.
//
// # Layer
//
// Lobby-domain — lobby-specific context type depending on kernel/ctx.
package ctx
