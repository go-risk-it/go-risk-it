package service

import "github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"

// RegionUpdate records a single region ownership or troop change produced by a
// performer or advancer.
type RegionUpdate struct {
	RegionID  string
	NewOwner  string
	NewTroops int64
}

// CardDelta records cards gained or lost by a single player during a move or
// advance step.
type CardDelta struct {
	PlayerUserID string
	Gained       []snapshot.CardState
	Lost         []int64
}

// MissionChange records a mission reassignment for a single player (e.g. after
// the original target is eliminated).
type MissionChange struct {
	PlayerUserID string
	NewMission   snapshot.PlayerMission
}

// DeckDelta records changes to the available (unowned) card deck during a move
// or advance step. Drawn lists card IDs removed from the deck (given to a
// player), Returned lists full CardState values added back (played cards
// returning to the deck).
type DeckDelta struct {
	Drawn    []int64
	Returned []snapshot.CardState
}

// MoveEffect is the aggregate side-effect bundle produced by a performer. It
// captures every observable state change so downstream consumers can diff
// without re-querying the database.
type MoveEffect struct {
	RegionUpdates []RegionUpdate
	CardDeltas    []CardDelta
	Missions      []MissionChange
	UpdatedPhase  snapshot.PhaseState
	DeckDelta     DeckDelta
	// EliminatedUserID is set by conquer performer when the defender is
	// eliminated (owns exactly one region being conquered). Empty for all
	// other moves.
	EliminatedUserID string
	// DeployableDelta is the change to DeployableTroops during a deploy move.
	// Negative values indicate troops deployed from the pool. Zero for all
	// other moves.
	DeployableDelta int64
}

// AdvanceEffect is the side-effect bundle produced by an advancer after a phase
// transition. It captures phase changes and any card movements triggered by the
// advance (e.g. card grants at end of turn).
type AdvanceEffect struct {
	NewPhase   snapshot.PhaseState
	TurnEnded  bool
	CardDeltas []CardDelta
	DeckDelta  DeckDelta
}
