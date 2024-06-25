package attack

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
)

type Move struct {
	AttackingRegionID string
	DefendingRegionID string
	TroopsInSource    int64
	TroopsInTarget    int64
	AttackingTroops   int64
}

type Service interface {
	service.Service[Move]
}

type ServiceImpl struct {
	regionService region.Service
}

var _ Service = &ServiceImpl{}

func NewService(regionService region.Service) *ServiceImpl {
	return &ServiceImpl{
		regionService: regionService,
	}
}

func (s *ServiceImpl) MustAdvanceQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	game *sqlc.Game,
) bool {
	return false
}

func (s *ServiceImpl) PerformQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	_ *sqlc.Game,
	move Move,
) error {
	ctx.Log().Infow("performing attack move", "move", move)

	attackingRegion, err := s.regionService.GetRegionQ(ctx, querier, move.AttackingRegionID)
	if err != nil {
		return fmt.Errorf("unable to get attacking region: %w", err)
	}

	defendingRegion, err := s.regionService.GetRegionQ(ctx, querier, move.DefendingRegionID)
	if err != nil {
		return fmt.Errorf("unable to get defending region: %w", err)
	}

	if attackingRegion.UserID != ctx.UserID() {
		return fmt.Errorf("attacking region is not owned by player")
	}

	if defendingRegion.UserID == ctx.UserID() {
		return fmt.Errorf("cannot attack your own region")
	}

	if move.AttackingTroops < 1 {
		return fmt.Errorf("at least one troop is required to attack")
	}

	if attackingRegion.Troops <= move.AttackingTroops {
		return fmt.Errorf("attacking region does not have enough troops")
	}

	if defendingRegion.Troops < 1 {
		ctx.Log().Errorw(
			"attempting to attack a region with no troops",
			"region",
			defendingRegion.ExternalReference,
		)

		return fmt.Errorf("defending region does not have enough troops")
	}

	if attackingRegion.Troops != move.TroopsInSource {
		return fmt.Errorf("attacking region doesn't have the declared number of troops")
	}

	if defendingRegion.Troops != move.TroopsInTarget {
		return fmt.Errorf("defending region doesn't have the declared number of troops")
	}

	return nil
}
