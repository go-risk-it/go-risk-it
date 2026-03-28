// Package rest defines the HTTP route table for game operations.
//
// [ProvideRoutes] returns typed [route.Route] values for game creation,
// game summary, move submission (deploy, attack, conquer, reinforce, cards),
// phase advancement, and the game WebSocket upgrade endpoint. Routes are
// contributed to the mux via fx group injection.
//
// # Layer
//
// Web — REST route registration for game endpoints.
package rest
