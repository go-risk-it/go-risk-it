package service

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

type Performer[T any] interface {
	PerformQ(ctx ctx.GameContext, querier db.Querier, move T) (any, error)
}

type Advancer interface {
	AdvanceQ(
		ctx ctx.GameContext,
		querier db.Querier,
		targetPhase sqlc.GamePhaseType,
		performResult any,
	) error
}

type PhaseWalker interface {
	WalkQ(
		ctx ctx.GameContext,
		querier db.Querier,
		voluntaryAdvancement bool,
	) (sqlc.GamePhaseType, error)
}

type Service[T any] interface {
	Performer[T]
	PhaseWalker
	Advancer
	PhaseType() sqlc.GamePhaseType
}
