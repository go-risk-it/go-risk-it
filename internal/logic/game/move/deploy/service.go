package deploy

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type Move struct {
	RegionID      string
	CurrentTroops int64
	DesiredTroops int64
}

type MoveResult struct{}

type Service interface {
	service.Service[Move, *MoveResult]
	GetDeployableTroops(ctx ctx.GameContext) (int64, error)
	GetDeployableTroopsQ(ctx ctx.GameContext, querier db.Querier) (int64, error)
}

type ServiceImpl struct {
	querier       db.Querier
	gameService   state.Service
	phaseService  phase.Service
	playerService player.Service
	regionService region.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
	gameService state.Service,
	phaseService phase.Service,
	playerService player.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		querier:       querier,
		gameService:   gameService,
		phaseService:  phaseService,
		playerService: playerService,
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
