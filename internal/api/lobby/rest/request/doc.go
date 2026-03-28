// Package request defines the JSON request bodies for lobby REST endpoints.
//
// Each struct maps to one lobby action (create, join, start) and implements
// a Validate method that checks required fields before the request reaches
// the logic layer.
//
// # Layer
//
// API — inbound REST DTOs with self-validation.
package request
