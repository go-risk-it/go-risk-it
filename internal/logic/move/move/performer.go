package move

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

type Move[T any] struct {
	UserID  string
	GameID  int64
	Payload T
}

type Performer[T any] interface {
	PerformQ(
		ctx context.Context,
		querier db.Querier,
		game *sqlc.Game,
		move Move[T],
	) error
}

type Service[T any] interface {
	Performer[T]
	ValidatePhase(game *sqlc.Game) bool
	MustAdvanceQ(
		ctx context.Context,
		querier db.Querier,
		game *sqlc.Game,
	) bool
}
