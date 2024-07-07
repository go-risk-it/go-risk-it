package deploy

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type Move struct {
	RegionID      string
	CurrentTroops int64
	DesiredTroops int64
}

type Service interface {
	service.Service[Move]
	GetDeployableTroops(ctx ctx.GameContext) (int64, error)
	GetDeployableTroopsQ(ctx ctx.GameContext, querier db.Querier) (int64, error)
}

type ServiceImpl struct {
	querier       db.Querier
	gameService   state.Service
	playerService player.Service
	regionService region.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
	gameService state.Service,
	playerService player.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		querier:       querier,
		gameService:   gameService,
		playerService: playerService,
		regionService: regionService,
	}
}

func (s *ServiceImpl) PerformQ(ctx ctx.MoveContext, querier db.Querier, move Move) error {
	ctx.Log().Infow("performing deploy move", "move", move)

	deployableTroops, err := s.GetDeployableTroopsQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to get deployable troops: %w", err)
	}

	troops := move.DesiredTroops - move.CurrentTroops
	if deployableTroops < troops {
		return fmt.Errorf("not enough deployable troops")
	}

	thisRegion, err := s.regionService.GetRegionQ(ctx, querier, move.RegionID)
	if err != nil {
		return fmt.Errorf("failed to get region: %w", err)
	}

	if thisRegion.UserID != ctx.UserID() {
		return fmt.Errorf("region is not owned by player")
	}

	if thisRegion.Troops != move.CurrentTroops {
		return fmt.Errorf("region has different number of troops than declared")
	}

	if err := s.executeDeploy(ctx, querier, thisRegion, troops); err != nil {
		return fmt.Errorf("failed to execute deploy: %w", err)
	}

	return nil
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

func (s *ServiceImpl) executeDeploy(
	ctx ctx.MoveContext,
	querier db.Querier,
	region *sqlc.GetRegionsByGameRow,
	troops int64,
) error {
	ctx.Log().Infow(
		"executing deploy",
		"region",
		region.ExternalReference,
		"troops",
		troops,
	)

	if err := s.decreaseDeployableTroopsQ(ctx, querier, troops); err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegion(ctx, querier, region.ID, troops); err != nil {
		return fmt.Errorf("failed to increase region troops: %w", err)
	}

	ctx.Log().Infow(
		"deploy executed successfully",
		"region",
		region.ExternalReference,
		"troops",
		troops,
	)

	return nil
}

func (s *ServiceImpl) decreaseDeployableTroopsQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	troops int64,
) error {
	ctx.Log().Infow("decreasing deployable troops", "troops", troops)

	err := querier.DecreaseDeployableTroops(ctx, sqlc.DecreaseDeployableTroopsParams{
		ID:               ctx.GameID(),
		DeployableTroops: troops,
	})
	if err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	ctx.Log().Infow("decreased deployable troops", "troops", troops)

	return nil
}
