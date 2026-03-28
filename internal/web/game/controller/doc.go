// Package controller provides web-layer controllers for in-game operations.
//
// Each controller translates between API request types and logic-layer service
// calls for a specific game domain: board state, player state, moves, phases,
// advancements, cards, missions, and move history. Controllers are thin
// adapters that own no business logic themselves.
//
// # Layer
//
// Web — HTTP request handling and API-to-domain translation.
package controller
