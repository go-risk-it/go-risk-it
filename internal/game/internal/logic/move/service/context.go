package service

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
)

// WalkContext carries all data a phase walker needs to decide the next phase
// without issuing additional database queries. It is populated by the
// orchestration layer before calling Walk.
type WalkContext struct {
	Voluntary        bool
	PrevSnapshot     *snapshot.GameSnapshot
	PrivateSnapshots map[string]*snapshot.PlayerPrivate
	Effect           MoveEffect
	CurrentUserID    string
}

// AdvanceContext carries all data an advancer needs to apply post-transition
// logic (e.g. troop grants, card awards) without re-querying the database.
type AdvanceContext struct {
	ConqueredInTurn bool
	UpdatedRegions  []snapshot.RegionState
	CurrentUserID   string
	Continents      board.Continents
	Turn            int64                  // pre-computed from CachedGameState
	CurrentPhase    string                 // current phase type for turn-boundary detection
	Players         []snapshot.PlayerState // for dead-player skip in getNextTurn
	AvailableDeck   []snapshot.CardState   // cached deck for card draw without DB query
}
