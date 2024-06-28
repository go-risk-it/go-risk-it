package attack

import (
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/dice"
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

	HasConqueredQ(ctx ctx.MoveContext, querier db.Querier) bool
	CanContinueAttackingQ(ctx ctx.MoveContext, querier db.Querier) bool
}

type ServiceImpl struct {
	boardService  board.Service
	diceService   dice.Service
	regionService region.Service
}

var _ Service = &ServiceImpl{}

func NewService(
	boardService board.Service,
	diceService dice.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		boardService:  boardService,
		diceService:   diceService,
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

	if err := s.validate(ctx, attackingRegion, defendingRegion, move); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.perform(ctx, querier, attackingRegion, defendingRegion, move); err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	ctx.Log().Infow("attack executed successfully")

	return nil
}

func (s *ServiceImpl) perform(
	ctx ctx.MoveContext,
	querier db.Querier,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	attackDices := s.diceService.Roll(int(move.AttackingTroops))
	defenseDices := s.diceService.Roll(int(min(defendingRegion.Troops, 3)))

	ctx.Log().Infow("rolled dices", "attack", attackDices, "defense", defenseDices)

	casualties := computeCasualties(ctx, attackDices, defenseDices)

	ctx.Log().Infow("updating region troops")

	if err := s.regionService.UpdateTroopsInRegion(
		ctx,
		querier,
		attackingRegion.ID,
		-casualties.attacking,
	); err != nil {
		return fmt.Errorf("failed to decrease troops in attacking region: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegion(
		ctx,
		querier,
		defendingRegion.ID,
		-casualties.defending,
	); err != nil {
		return fmt.Errorf("failed to decrease troops in defending region: %w", err)
	}

	return nil
}

type casualties struct {
	attacking int64
	defending int64
}

func computeCasualties(ctx ctx.MoveContext, attackDices, defenseDices []int) *casualties {
	casualties := &casualties{}

	slices.SortFunc(attackDices, descending)
	slices.SortFunc(defenseDices, descending)

	matches := min(len(attackDices), len(defenseDices))
	for i := 0; i < matches; i++ {
		if attackDices[i] > defenseDices[i] {
			casualties.defending++
		} else {
			casualties.attacking++
		}
	}

	ctx.Log().Infow(
		"casualties",
		"attacking",
		casualties.attacking,
		"defending",
		casualties.defending)

	return casualties
}

func descending(a, b int) int {
	return b - a
}

func (s *ServiceImpl) validate(
	ctx ctx.MoveContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	ctx.Log().Infow("validating attack move", "move", move)

	if err := checkRegionOwnership(ctx, attackingRegion, defendingRegion); err != nil {
		return fmt.Errorf("region ownership check failed: %w", err)
	}

	if err := checkTroops(ctx, attackingRegion, defendingRegion, move); err != nil {
		return fmt.Errorf("troops check failed: %w", err)
	}

	areNeighbours, err := s.boardService.AreNeighbours(
		ctx,
		attackingRegion.ExternalReference,
		defendingRegion.ExternalReference,
	)
	if err != nil {
		return fmt.Errorf("unable to check if regions are neighbours: %w", err)
	}

	if !areNeighbours {
		return fmt.Errorf("attacking region cannot reach defending region")
	}

	ctx.Log().Infow("attack move validation passed", "move", move)

	return nil
}

func checkTroops(
	ctx ctx.MoveContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	ctx.Log().Infow("checking troops")

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

	if err := checkDeclaredValues(ctx, attackingRegion, defendingRegion, move); err != nil {
		return fmt.Errorf("declared values are invalid: %w", err)
	}

	ctx.Log().Infow("troops check passed")

	return nil
}

func checkRegionOwnership(
	ctx ctx.MoveContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
) error {
	ctx.Log().Infow("checking region ownership")

	if attackingRegion.UserID != ctx.UserID() {
		return fmt.Errorf("attacking region is not owned by player")
	}

	if defendingRegion.UserID == ctx.UserID() {
		return fmt.Errorf("cannot attack your own region")
	}

	ctx.Log().Infow("region ownership check passed")

	return nil
}

func checkDeclaredValues(
	ctx ctx.MoveContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	if attackingRegion.Troops != move.TroopsInSource {
		return fmt.Errorf("attacking region doesn't have the declared number of troops")
	}

	if defendingRegion.Troops != move.TroopsInTarget {
		return fmt.Errorf("defending region doesn't have the declared number of troops")
	}

	ctx.Log().Infow("declared values check passed")

	return nil
}

func (s *ServiceImpl) HasConqueredQ(ctx ctx.MoveContext, querier db.Querier) bool {
	return false
}

func (s *ServiceImpl) CanContinueAttackingQ(ctx ctx.MoveContext, querier db.Querier) bool {
	return false
}
