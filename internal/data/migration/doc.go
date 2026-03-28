// Package migration runs database schema migrations at startup.
//
// [Execute] creates the target PostgreSQL schema if it does not exist, then
// applies all pending migration files from the migrations/ directory using
// golang-migrate. It is called once per schema (game, lobby) during
// application initialization.
//
// # Layer
//
// Data — schema migration with no domain dependencies.
package migration
