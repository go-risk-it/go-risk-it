package service

import (
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
)

type Performer[T, R any] interface {
	Perform(
		ctx ctx.GameContext,
		querier db.Querier,
		move T,
		prev *snapshot.CachedGameState,
	) (R, MoveEffect, error)
}

type Advancer[R any] interface {
	Advance(
		ctx ctx.GameContext,
		querier db.Querier,
		targetPhase sqlc.GamePhaseType,
		performResult R,
		advCtx AdvanceContext,
	) (AdvanceEffect, error)
}

type PhaseWalker interface {
	Walk(wctx WalkContext) (sqlc.GamePhaseType, error)
}

type Service[T, R any] interface {
	Performer[T, R]
	PhaseWalker
	Advancer[R]
	PhaseType() sqlc.GamePhaseType
}
