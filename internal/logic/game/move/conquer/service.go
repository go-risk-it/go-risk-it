package conquer

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
)

type Move struct {
	Troops int64
}

type MoveResult struct{}

type Service interface {
	service.Service[Move, *MoveResult]
	GetPhaseState(ctx ctx.GameContext) (sqlc.GetConquerPhaseStateRow, error)
	GetPhaseStateQ(ctx ctx.GameContext, querier db.Querier) (sqlc.GetConquerPhaseStateRow, error)
}

type ServiceImpl struct {
	querier       db.Querier
	attackService attack.Service
	phaseService  phase.Service
	regionService region.Service
}

func NewService(
	querier db.Querier,
	attackService attack.Service,
	phaseService phase.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		querier:       querier,
		attackService: attackService,
		phaseService:  phaseService,
		regionService: regionService,
	}
}

var _ Service = (*ServiceImpl)(nil)

func (s *ServiceImpl) GetPhaseState(ctx ctx.GameContext) (sqlc.GetConquerPhaseStateRow, error) {
	return s.GetPhaseStateQ(ctx, s.querier)
}

func (s *ServiceImpl) GetPhaseStateQ(
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

func (s *ServiceImpl) PhaseType() sqlc.PhaseType {
	return sqlc.PhaseTypeCONQUER
}
