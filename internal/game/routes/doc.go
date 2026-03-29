// Package routes defines the HTTP route table and REST controllers for game
// operations.
//
// Controllers translate between API request types and logic-layer service
// calls: game creation, move submission, and phase advancement. Routes are
// contributed to the mux via fx group injection. The [GameController]
// implements [commands.Handler] to serve as the cross-module game creation
// endpoint dispatched by the kernel router.
//
// # Layer
//
// Web — REST route registration and request handling for game endpoints.
package routes
