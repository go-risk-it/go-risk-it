package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
)

// MoveContext carries the non-effect metadata needed by buildPersistenceEffect
// to construct the MoveLogEntry. It is populated by the orchestration pipeline
// from the game context and move-specific data.
type MoveContext struct {
	gameID    int64
	userID    string
	phaseType sqlc.GamePhaseType
	moveData  any
	result    any
}

// PersistenceEffect is a pure data container carrying all database writes needed
// to persist a move's side-effects. It groups writes semantically into 6 categories
// ordered by foreign-key dependencies:
//
//  1. MoveLog — references game + player (base FKs)
//  2. MoveExecution — region mutations
//     (troops, ownership, deployable, card unlinks, region bonuses)
//  3. Elimination — cascade data (transfer cards, reassign missions, delete spurious)
//  4. CardDraw — independent card ownership changes
//  5. PhaseTransition — creates new FK-referenced rows (phase, conquer_phase, deploy_phase)
//  6. GameConclusion — updates game row (winner), semantically final
//
// All groups are optional pointers. A nil group means no writes needed in that
// category. Persist() skips nil groups, making this a zero-overhead no-op when
// no writes are required.
//
// Type design contract:
//   - Uses internal DB IDs (int64) for all region/player references since Persist()
//     operates with sqlc querier
//   - MoveExecution is grouped by kind (not flat list): RegionTroopUpdates,
//     OwnershipChanges, DeployableDelta, CardUnlinks, RegionBonuses
//   - EliminationEffect carries UserID (string), Persist() resolves to internal ID
//   - GameConclusion.WinnerUserID same pattern — carries UserID, Persist() resolves
type PersistenceEffect struct {
	MoveLog         *MoveLogEntry
	MoveExecution   *MoveExecution
	Elimination     *EliminationEffect
	CardDraw        *CardDraw
	PhaseTransition *PhaseTransition
	GameConclusion  *GameConclusion
}

// MoveLogEntry carries all data needed for CreateMoveLog. The phase type is stored
// as a direct string field (not a subquery reference), matching the T3 SQL change.
type MoveLogEntry struct {
	GameID    int64
	UserID    string
	PhaseType string
	MoveData  []byte
	Result    []byte
}

// MoveExecution groups all region-level and player-level mutations produced by a
// move. Each field represents a different kind of write operation:
//   - RegionTroopUpdates: troops added/removed from regions (signed deltas)
//   - OwnershipChanges: regions changing ownership (conquest)
//   - DeployableDelta: player's deployable troops change (single player, signed delta)
//   - CardUnlinks: card IDs to unlink from their owner (card plays)
//   - RegionBonuses: troops granted to multiple regions (territory bonuses)
type MoveExecution struct {
	RegionTroopUpdates []RegionTroopUpdate
	OwnershipChanges   []OwnershipChange
	DeployableDelta    *DeployableDelta
	CardUnlinks        []int64
	RegionBonuses      *RegionBonuses
}

// RegionTroopUpdate represents a signed delta applied to a single region's troops.
// Positive Delta increases troops, negative Delta decreases troops.
type RegionTroopUpdate struct {
	RegionID int64
	Delta    int64
}

// OwnershipChange represents a region conquest. The OldOwnerInternalID is the
// internal player ID returned by UpdateRegionOwner (needed for elimination checks).
type OwnershipChange struct {
	RegionID           int64
	NewOwnerUserID     string
	OldOwnerInternalID int64
}

// DeployableDelta represents a change to a single player's deployable troops.
// The Delta is signed: positive for increases, negative for decreases.
type DeployableDelta struct {
	PlayerID int64
	Delta    int64
}

// RegionBonuses represents troops granted to multiple regions simultaneously
// (e.g., territory bonuses at turn start).
type RegionBonuses struct {
	TroopsPerRegion int64
	RegionIDs       []int64
}

// EliminationEffect encapsulates the 3-write cascade triggered by player elimination:
//  1. TransferCardsOwnership (eliminated player's cards → conqueror)
//  2. ReassignMissions (missions targeting eliminated player → TwentyFourTerritories)
//  3. DeleteSpuriousEliminatePlayerMissions (missions targeting eliminated player)
//
// The EliminatedUserID is a string (external user ID). Persist() resolves it to
// the internal player ID via GetPlayerByUserId before executing the cascade.
type EliminationEffect struct {
	EliminatedUserID string
	ConquerorUserID  string
}

// CardDraw represents a single card drawn from the deck to a player.
type CardDraw struct {
	CardID int64
	UserID string
}

// PhaseTransition captures all data needed to insert a new phase row and optional
// sub-phase rows (conquer_phase, deploy_phase). The Turn and Players list are
// pre-computed by the advancer (including dead-player skip logic).
//
// ConquerData and DeployData are optional pointers — present only when transitioning
// to CONQUER or DEPLOY phases respectively.
type PhaseTransition struct {
	Turn      int64
	PhaseType string
	Players   []PlayerRef
	// ConquerData is present when transitioning to CONQUER phase.
	ConquerData *ConquerData
	// DeployData is present when transitioning to DEPLOY phase.
	DeployData *DeployData
}

// PlayerRef carries both UserID and InternalID for a player. This dual reference
// allows Persist() to use the appropriate ID for each write without additional queries.
type PlayerRef struct {
	UserID     string
	InternalID int64
}

// ConquerData holds the source/target region names (external references) and
// minimum troop count needed to populate the conquer_phase table row.
// The InsertConquerPhase SQL resolves region names to IDs via subquery.
type ConquerData struct {
	SourceRegionName string
	TargetRegionName string
	MinTroops        int64
}

// DeployData holds the deployable troop count needed to populate
// the deploy_phase table row.
type DeployData struct {
	DeployableTroops int64
}

// GameConclusion captures the winner's UserID when a game ends (mission accomplished
// or domination). Persist() resolves the UserID to an internal player ID via
// GetPlayerByUserId before calling AssignGameWinner.
type GameConclusion struct {
	WinnerUserID string
}

// CardState represents a single card in a player's hand. This is a snapshot type
// re-declared here to avoid depending on the api/snapshot package from the logic
// layer. It matches snapshot.CardState structurally.
type CardState struct {
	ID     int64
	Type   snapshot.CardType
	Region string
}
