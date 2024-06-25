package deploy

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	service2 "github.com/go-risk-it/go-risk-it/internal/logic/game/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"go.uber.org/zap"
)

type Move struct {
	RegionID      string
	CurrentTroops int64
	DesiredTroops int64
}

type Service interface {
	service.Service[Move]
}

type ServiceImpl struct {
	log           *zap.SugaredLogger
	querier       db.Querier
	gameService   service2.Service
	playerService player.Service
	regionService region.Service
}

var _ Service = &ServiceImpl{}

func NewService(
	que db.Querier,
	log *zap.SugaredLogger,
	gameService service2.Service,
	playerService player.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		querier:       que,
		log:           log,
		gameService:   gameService,
		playerService: playerService,
		regionService: regionService,
	}
}

func (s *ServiceImpl) MustAdvanceQ(
	_ ctx.MoveContext,
	_ db.Querier,
	game *sqlc.Game,
) bool {
	return game.DeployableTroops == 0
}

func (s *ServiceImpl) PerformQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	game *sqlc.Game,
	move Move,
) error {
	ctx.Log().Infow("performing deploy move", "move", move)

	troops := move.DesiredTroops - move.CurrentTroops
	if game.DeployableTroops < troops {
		return fmt.Errorf("not enough deployable troops")
	}

	regionState, err := s.getRegion(ctx, querier, move.RegionID)
	if err != nil {
		return fmt.Errorf("failed to get region: %w", err)
	}

	if regionState.Troops != move.CurrentTroops {
		return fmt.Errorf("region has different number of troops than declared")
	}

	err = s.executeDeploy(ctx, querier, game, regionState, troops)
	if err != nil {
		return fmt.Errorf("failed to execute deploy: %w", err)
	}

	return nil
}

func (s *ServiceImpl) executeDeploy(
	ctx ctx.MoveContext,
	querier db.Querier,
	game *sqlc.Game,
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

	err := s.gameService.DecreaseDeployableTroopsQ(ctx, querier, game, troops)
	if err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	err = s.regionService.IncreaseTroopsInRegion(ctx, querier, region.ID, troops)
	if err != nil {
		return fmt.Errorf("failed to increase region troops: %w", err)
	}

	return nil
}

func (s *ServiceImpl) getRegion(
	ctx ctx.MoveContext,
	querier db.Querier,
	region string,
) (*sqlc.GetRegionsByGameRow, error) {
	result, err := s.regionService.GetRegionQ(ctx, querier, ctx.GameID(), region)
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	if result.UserID != ctx.UserID() {
		return nil, fmt.Errorf("region is not owned by player")
	}

	return result, nil
}
