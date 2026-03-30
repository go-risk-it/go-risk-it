package orchestration

import (
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
)

type Orchestrator[T, R any] interface {
	OrchestrateMove(ctx gamectx.GameContext, move T) error
}

// moveOutcome carries the committed transaction result for post-commit event emission.
type moveOutcome[R any] struct {
	targetPhase sqlc.GamePhaseType
	gameOver    bool
	result      R
	moveLog     sqlc.GameMoveLog
	turn        int64
}
