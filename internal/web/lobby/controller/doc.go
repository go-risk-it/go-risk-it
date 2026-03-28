// Package controller provides web-layer controllers for lobby operations.
//
// Each controller translates between API request types and logic-layer
// service calls: lobby creation, player management (join, list), game
// start, and lobby state queries. Controllers are thin adapters that
// own no business logic themselves.
//
// # Layer
//
// Web — HTTP request handling and API-to-domain translation for lobbies.
package controller
