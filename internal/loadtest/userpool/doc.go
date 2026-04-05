// Package userpool manages a reusable pool of authenticated Supabase users
// for load testing. Instead of creating fresh users per game (which overwhelms
// Supabase auth at high churn rates), users are signed up lazily and reused
// across games with a configurable per-user concurrency limit.
//
// Thread-safe for concurrent Acquire/Release from multiple game goroutines.
package userpool
