package region

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region/assignment"
)

var (
	ErrNoPlayers                 = errors.New("no players provided")
	ErrPlayersFromDifferentGames = errors.New("players from different games")
)

type Service interface {
	CreateRegions(
		ctx ctx.UserContext,
		querier db.Querier,
		players []sqlc.Player,
		regions []string,
	) error

	GetRegionQ(
		ctx ctx.MoveContext,
		querier db.Querier,
		region string,
	) (*sqlc.GetRegionsByGameRow, error)
	GetRegions(ctx ctx.GameContext) ([]sqlc.GetRegionsByGameRow, error)
	GetRegionsQ(ctx ctx.GameContext, querier db.Querier) ([]sqlc.GetRegionsByGameRow, error)
	IncreaseTroopsInRegion(
		ctx ctx.MoveContext,
		querier db.Querier,
		regionID int64,
		troops int64,
	) error
}
type ServiceImpl struct {
	querier           db.Querier
	assignmentService assignment.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(querier db.Querier, assignmentService assignment.Service) *ServiceImpl {
	return &ServiceImpl{querier: querier, assignmentService: assignmentService}
}

func (s *ServiceImpl) CreateRegions(
	ctx ctx.UserContext,
	querier db.Querier,
	players []sqlc.Player,
	regions []string,
) error {
	ctx.Log().Infow("creating regions", "players_size", len(players), "regions_size", len(regions))

	if len(players) == 0 {
		return ErrNoPlayers
	}

	gameID := players[0].GameID

	for _, player := range players {
		if player.GameID != gameID {
			return ErrPlayersFromDifferentGames
		}
	}

	regionToPlayer := s.assignmentService.AssignRegionsToPlayers(players, regions)
	regionsParams := make([]sqlc.InsertRegionsParams, 0, len(regionToPlayer))

	for _, region := range regions {
		regionsParams = append(regionsParams, sqlc.InsertRegionsParams{
			ExternalReference: region,
			PlayerID:          regionToPlayer[region].ID,
			Troops:            3,
		})
	}

	if _, err := querier.InsertRegions(ctx, regionsParams); err != nil {
		return fmt.Errorf("failed to insert regions: %w", err)
	}

	ctx.Log().Infow("created regions", "players", players, "regions", regions)

	return nil
}

func (s *ServiceImpl) GetRegions(
	ctx ctx.GameContext,
) ([]sqlc.GetRegionsByGameRow, error) {
	return s.GetRegionsQ(ctx, s.querier)
}

func (s *ServiceImpl) GetRegionsQ(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.GetRegionsByGameRow, error) {
	ctx.Log().Infow("fetching regions")

	regions, err := querier.GetRegionsByGame(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}

	ctx.Log().Debugw("got regions", "regions", len(regions))

	return regions, nil
}

func (s *ServiceImpl) GetRegionQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	region string,
) (*sqlc.GetRegionsByGameRow, error) {
	ctx.Log().Infow("fetching region", "region", region)

	regions, err := s.GetRegionsQ(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}

	result := extractRegionFrom(region, regions)
	if result == nil {
		return nil, fmt.Errorf("region is not in game")
	}

	return result, nil
}

func extractRegionFrom(
	region string,
	regions []sqlc.GetRegionsByGameRow,
) *sqlc.GetRegionsByGameRow {
	for _, r := range regions {
		if r.ExternalReference == region {
			return &r
		}
	}

	return nil
}

func (s *ServiceImpl) IncreaseTroopsInRegion(
	ctx ctx.MoveContext,
	querier db.Querier,
	regionID int64,
	troops int64,
) error {
	ctx.Log().Infow("increasing region troops", "region_id", regionID, "troops", troops)

	err := querier.IncreaseRegionTroops(ctx, sqlc.IncreaseRegionTroopsParams{
		ID:     regionID,
		Troops: troops,
	})
	if err != nil {
		return fmt.Errorf("failed to increase region troops: %w", err)
	}

	ctx.Log().Infow("increased region troops", "region_id", regionID, "troops", troops)

	return nil
}
