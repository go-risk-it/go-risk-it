// Package rest defines the HTTP route table for lobby operations.
//
// [ProvideRoutes] returns typed [route.Route] values for lobby creation,
// lobby summary, join, game start, and the lobby WebSocket upgrade endpoint.
// Routes are contributed to the mux via fx group injection.
//
// # Layer
//
// Web — REST route registration for lobby endpoints.
package rest
