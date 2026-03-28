// Package request defines the JSON request bodies for game REST endpoints.
//
// Each struct maps to one game action (deploy, attack, conquer, reinforce,
// cards, advance phase, create game) and implements a Validate method that
// checks required fields and value bounds before the request reaches the
// logic layer.
//
// # Layer
//
// API — inbound REST DTOs with self-validation.
package request
