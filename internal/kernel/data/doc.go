// Package data provides shared database primitives for transaction management.
//
// It defines the [Transactable] generic interface that domain querier types
// must satisfy to participate in [InTransaction] and
// [InTransactionWithIsolation]. These functions execute a callback within a
// PostgreSQL transaction, with automatic retry on serialization failures
// (40001) and deadlocks (40P01), OTel span tracing, and commit/rollback
// lifecycle management. [TxBeginner] and [TxSupport] provide embeddable
// transaction-starting capabilities for domain querier structs.
//
// # Layer
//
// Kernel — database transaction infrastructure with no domain dependencies.
package data
