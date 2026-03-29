package region

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/region/assignment"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

var (
	ErrNoPlayers                 = domainerrors.NewValidationError("no players provided")
	ErrPlayersFromDifferentGames = domainerrors.NewValidationError("players from different games")
)

type Service interface {
	CreateRegions(
		ctx context.Context,
		querier db.Querier,
		players []sqlc.GamePlayer,
		regions []string,
	) error
	GetRegion(
		ctx ctx.GameContext,
		querier db.Querier,
		region string,
	) (*sqlc.GetRegionsByGameRow, error)
	GetRegions(ctx ctx.GameContext) ([]sqlc.GetRegionsByGameRow, error)
	GetRegionsWithQuerier(
		ctx ctx.GameContext,
		querier db.Querier,
	) ([]sqlc.GetRegionsByGameRow, error)
	GetPlayerRegions(ctx ctx.GameContext, querier db.Querier) ([]sqlc.GetRegionsByGameRow, error)
	GetRegionsControlledByPlayer(
		ctx ctx.GameContext,
		querier db.Querier,
		playerID int64,
	) ([]sqlc.GameRegion, error)
	UpdateTroopsInRegion(
		ctx ctx.GameContext,
		querier db.Querier,
		region *sqlc.GetRegionsByGameRow,
		troopsToAdd int64,
	) error
	UpdateRegionOwner(
		ctx ctx.GameContext,
		querier db.Querier,
		region *sqlc.GetRegionsByGameRow) (int64, error)
}

type service struct {
	querier           db.Querier
	assignmentService assignment.Service
}

func (s *service) GetRegionsControlledByPlayer(
	ctx ctx.GameContext,
	querier db.Querier,
	playerID int64,
) ([]sqlc.GameRegion, error) {
	return querier.GetRegionsByPlayer(ctx, playerID)
}

var _ Service = (*service)(nil)

func NewService(querier db.Querier, assignmentService assignment.Service) Service {
	return &service{querier: querier, assignmentService: assignmentService}
}

func (s *service) CreateRegions(
	ctx context.Context,
	querier db.Querier,
	players []sqlc.GamePlayer,
	regions []string,
) error {
	slog.InfoContext(ctx, "creating regions",
		"players_size", len(players), "regions_size", len(regions))

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

	slog.InfoContext(ctx, "created regions", "players", players, "regions", regions)

	return nil
}

func (s *service) GetRegions(
	ctx ctx.GameContext,
) ([]sqlc.GetRegionsByGameRow, error) {
	return s.GetRegionsWithQuerier(ctx, s.querier)
}

func (s *service) GetRegionsWithQuerier(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.GetRegionsByGameRow, error) {
	slog.DebugContext(ctx, "fetching regions")

	regions, err := querier.GetRegionsByGame(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}

	slog.DebugContext(ctx, "got regions", "regions", len(regions))

	return regions, nil
}

func (s *service) GetPlayerRegions(
	ctx ctx.GameContext,
	querier db.Querier,
) ([]sqlc.GetRegionsByGameRow, error) {
	regions, err := s.GetRegionsWithQuerier(ctx, querier)
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

func (s *service) GetRegion(
	ctx ctx.GameContext,
	querier db.Querier,
	region string,
) (*sqlc.GetRegionsByGameRow, error) {
	slog.DebugContext(ctx, "fetching region", "region", region)

	regions, err := s.GetRegionsWithQuerier(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}

	result := extractRegionFrom(region, regions)
	if result == nil {
		return nil, domainerrors.NewNotFoundError("region is not in game")
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

func (s *service) UpdateTroopsInRegion(
	ctx ctx.GameContext,
	querier db.Querier,
	region *sqlc.GetRegionsByGameRow,
	troopsToAdd int64,
) error {
	if troopsToAdd == 0 {
		slog.DebugContext(ctx, "no troops to update")

		return nil
	}

	action := "increas"
	if troopsToAdd < 0 {
		action = "decreas"
	}

	slog.DebugContext(
		ctx,
		action+"ing troops in region "+region.ExternalReference,
		"troopsToAdd", troopsToAdd,
	)

	err := querier.IncreaseRegionTroops(ctx, sqlc.IncreaseRegionTroopsParams{
		ID:     region.ID,
		Troops: troopsToAdd,
	})
	if err != nil {
		return fmt.Errorf("failed to %se region troops: %w", action, err)
	}

	slog.DebugContext(ctx, action+"ed region troops")

	return nil
}

func (s *service) UpdateRegionOwner(
	ctx ctx.GameContext,
	querier db.Querier,
	region *sqlc.GetRegionsByGameRow,
) (int64, error) {
	slog.DebugContext(ctx, "updating region owner", "region", region.ExternalReference)

	oldOwnerPlayerID, err := querier.UpdateRegionOwner(ctx, sqlc.UpdateRegionOwnerParams{
		NewOwnerUserID:    ctx.UserID(),
		GameID:            ctx.GameID(),
		ConqueredRegionID: region.ID,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to update region owner: %w", err)
	}

	slog.DebugContext(ctx, "updated region owner", "region", region.ExternalReference)

	return oldOwnerPlayerID, nil
}
