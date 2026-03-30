// Package route provides the Route type, module-agnostic route constructors,
// and shared HTTP plumbing for the REST API.
//
// Three constructor levels encode authentication and context requirements:
//
//   - [Public] — no auth, plain handler.
//   - [Authed] — JWT auth, plain handler.
//   - [Domain] / [DomainWS] — JWT auth, domain context builder + typed handler.
//
// Module-specific constructors (Game, GameWS, Lobby, LobbyWS) in their
// respective route packages delegate to [Domain] / [DomainWS] with
// concrete context builders.
//
// All constructors wrap handlers with [WrapErrors], which maps domain
// errors to HTTP status codes and records them on the OTel span.
//
// # Layer
//
// Web — route metadata, handler types, and error translation.
package route
