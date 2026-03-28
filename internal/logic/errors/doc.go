// Package domainerrors defines typed domain errors with category-based HTTP
// status mapping. All domain errors implement the [Categorizable] interface,
// allowing error handlers to map errors to appropriate HTTP status codes
// (400, 401, 403, 404, 409) without type-switching on concrete types.
//
// [DomainError] is the single error type. Constructor functions
// (NewValidationError, NewConflictError, etc.) set the category, and Wrap
// variants preserve a cause chain compatible with [errors.Is] and
// [errors.As].
//
// # Layer
//
// Shared — domain error types used across all logic packages.
package domainerrors
