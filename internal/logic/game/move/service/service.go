package service

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

type Performer[T, R any] interface {
	Perform(ctx ctx.GameContext, querier db.Querier, move T) (R, error)
}

type Advancer[R any] interface {
	Advance(
		ctx ctx.GameContext,
		querier db.Querier,
		targetPhase sqlc.GamePhaseType,
		performResult R,
	) error
}

type PhaseWalker interface {
	Walk(
		ctx ctx.GameContext,
		querier db.Querier,
		voluntaryAdvancement bool,
	) (sqlc.GamePhaseType, error)
}

type Service[T, R any] interface {
	Performer[T, R]
	PhaseWalker
	Advancer[R]
	PhaseType() sqlc.GamePhaseType
}
