// Package routes defines the HTTP route table and REST controllers for lobby
// operations.
//
// Controllers translate between API request types and logic-layer service
// calls: lobby creation, player management (join, list), and game start.
// The [StartController] dispatches game creation through the kernel router
// rather than importing game packages directly. Routes are contributed to
// the mux via fx group injection.
//
// # Layer
//
// Web — REST route registration and request handling for lobby endpoints.
package routes
