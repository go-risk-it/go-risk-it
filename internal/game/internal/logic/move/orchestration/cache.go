package orchestration

import (
	apisnapshot "github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/state"
)

// MapSnapshotPhaseToSqlc maps a snapshot PhaseType to the corresponding sqlc
// GamePhaseType. Panics on unknown phase — this is an invariant violation
// (same pattern as BuildNewState exhaustive switches).
func MapSnapshotPhaseToSqlc(phaseType apisnapshot.PhaseType) sqlc.GamePhaseType {
	switch phaseType {
	case apisnapshot.PhaseCards:
		return sqlc.GamePhaseTypeCARDS
	case apisnapshot.PhaseDeploy:
		return sqlc.GamePhaseTypeDEPLOY
	case apisnapshot.PhaseAttack:
		return sqlc.GamePhaseTypeATTACK
	case apisnapshot.PhaseConquer:
		return sqlc.GamePhaseTypeCONQUER
	case apisnapshot.PhaseReinforce:
		return sqlc.GamePhaseTypeREINFORCE
	default:
		panic("unknown snapshot phase type: " + string(phaseType))
	}
}

// GameStateFromCache builds a state.Game from the cached game state. This
// replaces the GetGameStateWithQuerier DB call in the orchestration pipeline.
func GameStateFromCache(cached *apisnapshot.CachedGameState) *state.Game {
	return &state.Game{
		ID:           cached.PublicSnapshot.Game.ID,
		Turn:         cached.Turn,
		Phase:        MapSnapshotPhaseToSqlc(cached.PublicSnapshot.Phase.Type),
		WinnerUserID: cached.PublicSnapshot.Game.WinnerUserID,
	}
}
