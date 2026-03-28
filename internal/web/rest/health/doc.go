// Package health provides the GET /status endpoint for liveness and
// readiness checks.
//
// [New] configures a health check handler that verifies PostgreSQL
// connectivity and returns component metadata. The route is registered
// as a public (unauthenticated) endpoint.
//
// # Layer
//
// Web — health check endpoint.
package health
