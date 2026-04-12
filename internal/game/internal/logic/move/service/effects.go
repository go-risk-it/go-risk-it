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

// MoveEffect is the aggregate side-effect bundle produced by a performer. It
// captures every observable state change so downstream consumers can diff
// without re-querying the database.
type MoveEffect struct {
	RegionUpdates []RegionUpdate
	CardDeltas    []CardDelta
	Missions      []MissionChange
	UpdatedPhase  snapshot.PhaseState
}

// AdvanceEffect is the side-effect bundle produced by an advancer after a phase
// transition. It captures phase changes and any card movements triggered by the
// advance (e.g. card grants at end of turn).
type AdvanceEffect struct {
	NewPhase   snapshot.PhaseState
	TurnEnded  bool
	CardDeltas []CardDelta
}
