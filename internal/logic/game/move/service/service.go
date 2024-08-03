package service

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

type Performer[T any] interface {
	PerformQ(ctx ctx.MoveContext, querier db.Querier, move T) error
}

type Advancer[T any] interface {
	AdvanceQ(ctx ctx.MoveContext, querier db.Querier, targetPhase sqlc.PhaseType, move T) error
}

type PhaseWalker interface {
	Walk(ctx ctx.MoveContext, querier db.Querier) (sqlc.PhaseType, error)
}

type Service[T any] interface {
	Advancer[T]
	Performer[T]
	PhaseWalker
}
