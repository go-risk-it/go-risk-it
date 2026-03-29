// Package route provides the Route type, module-agnostic route constructors,
// and shared HTTP plumbing for the REST API.
//
// Two constructors encode authentication requirements at the type level:
//
//   - [Public] — no auth, plain handler.
//   - [Authed] — JWT auth, plain handler.
//
// Module-specific constructors (Game, GameWS, Lobby, LobbyWS) live in their
// respective module route packages (game/routes/, lobby/routes/).
//
// All constructors wrap handlers with [WrapErrors], which maps domain
// errors to HTTP status codes and records them on the OTel span.
//
// # Layer
//
// Web — route metadata, handler types, and error translation.
package route
