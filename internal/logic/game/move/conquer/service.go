package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

type Service interface {
	GetPhaseState(ctx ctx.GameContext) (sqlc.GetConquerPhaseStateRow, error)
	GetPhaseStateQ(ctx ctx.GameContext, querier db.Querier) (sqlc.GetConquerPhaseStateRow, error)
}

type ServiceImpl struct {
	querier db.Querier
}

func NewService(querier db.Querier) *ServiceImpl {
	return &ServiceImpl{querier: querier}
}

var _ Service = (*ServiceImpl)(nil)

func (s ServiceImpl) GetPhaseState(ctx ctx.GameContext) (sqlc.GetConquerPhaseStateRow, error) {
	return s.GetPhaseStateQ(ctx, s.querier)
}

func (s ServiceImpl) GetPhaseStateQ(
	ctx ctx.GameContext,
	querier db.Querier,
) (sqlc.GetConquerPhaseStateRow, error) {
	ctx.Log().Info("getting conquer phase state")

	conquerPhase, err := querier.GetConquerPhaseState(ctx, ctx.GameID())
	if err != nil {
		return sqlc.GetConquerPhaseStateRow{}, fmt.Errorf(
			"failed to get conquer phase state: %w",
			err,
		)
	}

	ctx.Log().Infow("got conquer phase state", "phase", conquerPhase)

	return conquerPhase, nil
}
