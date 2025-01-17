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
	CreateRegionsQ(
		ctx ctx.LogContext,
		querier db.Querier,
		players []sqlc.Player,
		regions []string,
	) error
	GetRegionQ(
		ctx ctx.GameContext,
		querier db.Querier,
		region string,
	) (*sqlc.GetRegionsByGameRow, error)
	GetRegions(ctx ctx.GameContext) ([]sqlc.GetRegionsByGameRow, error)
	GetRegionsQ(ctx ctx.GameContext, querier db.Querier) ([]sqlc.GetRegionsByGameRow, error)
	GetPlayerRegionsQ(ctx ctx.GameContext, querier db.Querier) ([]sqlc.GetRegionsByGameRow, error)
	GetRegionsControlledByPlayerQ(
		ctx ctx.GameContext,
		querier db.Querier,
		playerID int64,
	) ([]sqlc.Region, error)
	UpdateTroopsInRegionQ(
		ctx ctx.GameContext,
		querier db.Querier,
		region *sqlc.GetRegionsByGameRow,
		troopsToAdd int64,
	) error
	UpdateRegionOwnerQ(
		ctx ctx.GameContext,
		querier db.Querier,
		region *sqlc.GetRegionsByGameRow) error
}
type ServiceImpl struct {
	querier           db.Querier
	assignmentService assignment.Service
}

func (s *ServiceImpl) GetRegionsControlledByPlayerQ(
	ctx ctx.GameContext,
	querier db.Querier,
	playerID int64,
) ([]sqlc.Region, error) {
	return querier.GetRegionsByPlayer(ctx, playerID)
}

var _ Service = (*ServiceImpl)(nil)

func NewService(querier db.Querier, assignmentService assignment.Service) *ServiceImpl {
	return &ServiceImpl{querier: querier, assignmentService: assignmentService}
}

func (s *ServiceImpl) CreateRegionsQ(
	ctx ctx.LogContext,
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

func (s *ServiceImpl) GetPlayerRegionsQ(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.GetRegionsByGameRow, error) {
	regions, err := s.GetRegionsQ(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}

	return getPlayerRegions(ctx, regions), nil
}

func getPlayerRegions(
	ctx ctx.GameContext,
	regions []sqlc.GetRegionsByGameRow,
) []sqlc.GetRegionsByGameRow {
	result := make([]sqlc.GetRegionsByGameRow, 0)

	for _, region := range regions {
		if region.UserID == ctx.UserID() {
			result = append(result, region)
		}
	}

	return result
}

func (s *ServiceImpl) GetRegionQ(
	ctx ctx.GameContext,
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
		return nil, errors.New("region is not in game")
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

func (s *ServiceImpl) UpdateTroopsInRegionQ(
	ctx ctx.GameContext,
	querier db.Querier,
	region *sqlc.GetRegionsByGameRow,
	troopsToAdd int64,
) error {
	if troopsToAdd == 0 {
		ctx.Log().Infow("no troops to update")

		return nil
	}

	action := "increas"
	if troopsToAdd < 0 {
		action = "decreas"
	}

	ctx.Log().
		Infof("%sing troops in region %s by %d", action, region.ExternalReference, troopsToAdd)

	err := querier.IncreaseRegionTroops(ctx, sqlc.IncreaseRegionTroopsParams{
		ID:     region.ID,
		Troops: troopsToAdd,
	})
	if err != nil {
		return fmt.Errorf("failed to %se region troops: %w", action, err)
	}

	ctx.Log().Infof("%sed region troops", action)

	return nil
}

func (s *ServiceImpl) UpdateRegionOwnerQ(
	ctx ctx.GameContext,
	querier db.Querier,
	region *sqlc.GetRegionsByGameRow,
) error {
	ctx.Log().Infow("updating region owner", "region", region.ExternalReference)

	err := querier.UpdateRegionOwner(ctx, sqlc.UpdateRegionOwnerParams{
		UserID: ctx.UserID(),
		GameID: ctx.GameID(),
		ID:     region.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update region owner: %w", err)
	}

	ctx.Log().Infow("updated region owner", "region", region.ExternalReference)

	return nil
}
