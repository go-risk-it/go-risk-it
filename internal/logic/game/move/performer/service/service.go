package service

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
)

type Performer[T any] interface {
	PerformQ(ctx ctx.MoveContext, querier db.Querier, move T) error
}

type Service[T any] interface {
	Performer[T]
}
