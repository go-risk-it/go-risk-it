package service

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

type Performer[T, R any] interface {
	PerformQ(ctx ctx.MoveContext, querier db.Querier, move T) (R, error)
}

type Advancer[R any] interface {
	AdvanceQ(
		ctx ctx.MoveContext,
		querier db.Querier,
		targetPhase sqlc.PhaseType,
		performResult R,
	) error
}

type PhaseWalker interface {
	Walk(ctx ctx.MoveContext, querier db.Querier) (sqlc.PhaseType, error)
}

type Service[T, R any] interface {
	Performer[T, R]
	PhaseWalker
	Advancer[R]
}
