// Package db composes the sqlc-generated game queries with transaction support.
//
// [Querier] combines the sqlc.Querier query interface with [db.Transactable]
// so that game data access can participate in generic transactions managed by
// [db.InTransaction]. [Queries] is the concrete implementation backed by a
// pgx connection pool, with [WithTx] returning a transaction-scoped copy.
//
// # Layer
//
// Data — game-schema database access built on sqlc-generated code.
package db
