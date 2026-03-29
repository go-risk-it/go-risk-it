// Package ctx defines the game-scoped context type that enriches
// [kernelctx.UserContext] with a game ID. [GameContext] composes user identity,
// OTel tracing, and game-scoped metadata so that every layer of the game module
// has typed access to the active game. [WithGameID] constructs a GameContext
// from an authenticated user context.
//
// # Layer
//
// Game-domain — game-specific context type depending on kernel/ctx.
package ctx
