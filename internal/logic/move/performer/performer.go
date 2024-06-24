package performer

import (
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
)

type Performer[T any] interface {
	PerformQ(ctx riskcontext.MoveContext, querier db.Querier, game *sqlc.Game, move T) error
}

type Service[T any] interface {
	Performer[T]
	MustAdvanceQ(ctx riskcontext.MoveContext, querier db.Querier, game *sqlc.Game) bool
}
