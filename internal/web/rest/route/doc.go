// Package route provides typed route constructors and handler signatures
// for the REST API.
//
// Six constructors encode the authentication and context enrichment
// requirements at the type level:
//
//   - [Public] — no auth, plain handler.
//   - [Authed] — JWT auth, plain handler.
//   - [Game] — JWT auth, extracts {id} into GameContext.
//   - [Lobby] — JWT auth, extracts {id} into LobbyContext.
//   - [GameWS] — JWT auth + WebSocket flag, extracts {id} into GameContext.
//   - [LobbyWS] — JWT auth + WebSocket flag, extracts {id} into LobbyContext.
//
// All constructors wrap handlers with [WrapErrors], which maps domain
// errors to HTTP status codes and records them on the OTel span.
//
// # Layer
//
// Web — route metadata, handler types, and error translation.
package route
