package service

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

type Performer[T any] interface {
	PerformQ(ctx ctx.MoveContext, querier db.Querier, game *sqlc.Game, move T) error
}

type Service[T any] interface {
	Performer[T]
}
