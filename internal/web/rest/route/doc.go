// Package route provides the Route type, module-agnostic route constructors,
// and shared HTTP plumbing for the REST API.
//
// # Layer-1 constructors (low-level)
//
// These accept a [PlainHandler] or typed handler and return a [*Route]:
//
//   - [Public] — no auth, plain handler.
//   - [Authed] — JWT auth, plain handler.
//   - [Domain] / [DomainWS] — JWT auth, domain context builder + typed handler.
//
// # Layer-2 constructors (high-level)
//
// These compose layer-1 constructors with request decoding, response writing,
// and context extraction to eliminate per-route boilerplate:
//
//   - [CreateHandler] — Authed + decode Req + perform → Resp + 201 JSON.
//   - [QueryHandler] — Authed + query → Resp + 200 JSON.
//   - [DomainCommand] — Domain + decode Req + perform → 204.
//   - [DomainVoid] — Domain + perform → 204.
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
