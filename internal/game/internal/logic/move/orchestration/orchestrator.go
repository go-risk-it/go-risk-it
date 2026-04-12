package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
)

type Orchestrator[T, R any] interface {
	PhaseAdvancer // orchestrators can advance phases voluntarily

	OrchestrateMove(ctx gamectx.GameContext, move T) error
}

// moveOutcome carries the post-commit result for event emission and cache update.
// newState is produced by BuildNewState from enrichment data (MoveEffect + AdvanceEffect).
// prevRegions is captured from the pre-mutation state for headline detection.
type moveOutcome[R any] struct {
	targetPhase sqlc.GamePhaseType
	gameOver    bool
	result      R
	turn        int64
	newState    *snapshot.CachedGameState
	prevRegions []snapshot.RegionState
}
