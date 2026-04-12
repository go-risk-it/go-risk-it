package orchestration

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

// NewMoveContext creates a MoveContext for testing.
func NewMoveContext(
	gameID int64,
	userID string,
	phaseType sqlc.GamePhaseType,
	moveData any,
	result any,
) MoveContext {
	return MoveContext{
		gameID:    gameID,
		userID:    userID,
		phaseType: phaseType,
		moveData:  moveData,
		result:    result,
	}
}

// BuildPersistenceEffect is the exported test-only wrapper for buildPersistenceEffect.
func BuildPersistenceEffect(
	moveCtx MoveContext,
	moveEffect *moveservice.MoveEffect,
	advEffect *moveservice.AdvanceEffect,
	prevState *snapshot.CachedGameState,
	targetPhase sqlc.GamePhaseType,
	gameOver bool,
) *PersistenceEffect {
	return buildPersistenceEffect(moveCtx, moveEffect, advEffect, prevState, targetPhase, gameOver)
}
