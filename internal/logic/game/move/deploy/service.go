package deploy

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
)

type Move struct {
	RegionID      string `json:"regionId"`
	CurrentTroops int64  `json:"currentTroops"`
	DesiredTroops int64  `json:"desiredTroops"`
}

type MoveResult struct{}

type Service interface {
	service.Service[Move, *MoveResult]
	GetDeployableTroops(ctx ctx.GameContext) (int64, error)
	GetDeployableTroopsQ(ctx ctx.GameContext, querier db.Querier) (int64, error)
}

type ServiceImpl struct {
	querier       db.Querier
	phaseService  phase.Service
	regionService region.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
	phaseService phase.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		querier:       querier,
		phaseService:  phaseService,
		regionService: regionService,
	}
}

func (s *ServiceImpl) GetDeployableTroops(ctx ctx.GameContext) (int64, error) {
	return s.GetDeployableTroopsQ(ctx, s.querier)
}

func (s *ServiceImpl) GetDeployableTroopsQ(
	ctx ctx.GameContext,
	querier db.Querier,
) (int64, error) {
	ctx.Log().Infow("getting deployable troops")

	deployableTroops, err := querier.GetDeployableTroops(ctx, ctx.GameID())
	if err != nil {
		return 0, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	ctx.Log().Infow("got deployable troops", "troops", deployableTroops)

	return deployableTroops, nil
}

func (s *ServiceImpl) PhaseType() sqlc.PhaseType {
	return sqlc.PhaseTypeDEPLOY
}
