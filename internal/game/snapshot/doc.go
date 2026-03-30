// Package snapshot provides aggregated read models for efficient game state
// retrieval. [Service.GetPublicSnapshot] returns the full public state
// (game metadata, phase, board, players) and
// [Service.GetPrivateSnapshotsByUser] returns per-player private state
// (cards, mission). All queries execute sequentially on a single connection,
// avoiding the DB pool starvation caused by parallel fetcher goroutines.
//
// # Layer
//
// Game-support — aggregated game state read models.
//
// # Key Types
//
// [Service] is the primary interface providing GetPublicSnapshot and
// GetPrivateSnapshotsByUser for efficient single-connection reads.
//
// [PublicSnapshot] holds the full public game state: game metadata,
// phase-specific state, board regions, and player summaries.
//
// [PrivateSnapshot] holds per-player private state: cards and mission.
//
// [PhaseState] is a discriminated union holding phase-specific data,
// with only the pointer field matching Type set to non-nil.
//
// # Dependencies
//
//   - [db.Querier] for sequential game database queries on a single connection
package snapshot
