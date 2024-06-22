package deploy

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/move"
	"github.com/go-risk-it/go-risk-it/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/region"
	"go.uber.org/zap"
)

type MoveData struct {
	RegionID      string
	CurrentTroops int64
	DesiredTroops int64
}

type Service interface {
	move.Service[MoveData]
}

type ServiceImpl struct {
	log           *zap.SugaredLogger
	querier       db.Querier
	gameService   game.Service
	playerService player.Service
	regionService region.Service
}

func NewService(
	que db.Querier,
	log *zap.SugaredLogger,
	gameService game.Service,
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

func (s *ServiceImpl) ValidatePhase(game *sqlc.Game) bool {
	return game.Phase == sqlc.PhaseDEPLOY
}

func (s *ServiceImpl) MustAdvanceQ(
	_ context.Context,
	_ db.Querier,
	game *sqlc.Game,
) bool {
	return game.DeployableTroops == 0
}

func (s *ServiceImpl) PerformQ(
	ctx context.Context,
	querier db.Querier,
	move move.Move[MoveData],
	game *sqlc.Game,
) error {
	s.log.Infow(
		"performing deploy move",
		"gameID",
		move.GameID,
		"userID",
		move.UserID,
		"move",
		move,
	)

	troops := move.Payload.DesiredTroops - move.Payload.CurrentTroops
	if game.DeployableTroops < troops {
		return fmt.Errorf("not enough deployable troops")
	}

	regionState, err := s.getRegion(ctx, querier, move.GameID, move.Payload.RegionID, move.UserID)
	if err != nil {
		return fmt.Errorf("failed to get region: %w", err)
	}

	if regionState.Troops != move.Payload.CurrentTroops {
		return fmt.Errorf("region has different number of troops than declared")
	}

	err = s.executeDeploy(ctx, querier, game, regionState, troops)
	if err != nil {
		return fmt.Errorf("failed to execute deploy: %w", err)
	}

	return nil
}

func (s *ServiceImpl) executeDeploy(
	ctx context.Context,
	querier db.Querier,
	game *sqlc.Game,
	region *sqlc.GetRegionsByGameRow,
	troops int64,
) error {
	s.log.Infow(
		"deploying",
		"gameID",
		game.ID,
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
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	region string,
	userID string,
) (*sqlc.GetRegionsByGameRow, error) {
	result, err := s.regionService.GetRegionQ(ctx, querier, gameID, region)
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	if result.UserID != userID {
		return nil, fmt.Errorf("region is not owned by player")
	}

	return result, nil
}
