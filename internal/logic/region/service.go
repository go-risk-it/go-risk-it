package region

import (
	"context"
	"errors"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/data/db"
	sqlc "github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/region/assignment"
	"go.uber.org/zap"
)

var (
	ErrNoPlayers                 = errors.New("no players provided")
	ErrPlayersFromDifferentGames = errors.New("players from different games")
)

type Service interface {
	CreateRegions(
		ctx context.Context,
		querier db.Querier,
		players []sqlc.Player,
		regions []board.Region,
	) error
	GetRegions(ctx context.Context, gameID int64) ([]sqlc.GetRegionsByGameRow, error)
}
type ServiceImpl struct {
	log               *zap.SugaredLogger
	querier           db.Querier
	assignmentService assignment.Service
}

func NewRegionService(
	log *zap.SugaredLogger,
	querier db.Querier,
	assignmentService assignment.Service,
) *ServiceImpl {
	return &ServiceImpl{log: log, querier: querier, assignmentService: assignmentService}
}

func (s *ServiceImpl) CreateRegions(
	ctx context.Context,
	querier db.Querier,
	players []sqlc.Player,
	regions []board.Region,
) error {
	s.log.Infow("creating regions", "players", players, "regions", regions)

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
			ExternalReference: region.ExternalReference,
			PlayerID:          regionToPlayer[region].ID,
			Troops:            3,
		})
	}

	if _, err := querier.InsertRegions(ctx, regionsParams); err != nil {
		return fmt.Errorf("failed to insert regions: %w", err)
	}

	s.log.Infow("created regions", "players", players, "regions", regions)

	return nil
}

func (s *ServiceImpl) GetRegions(
	ctx context.Context,
	gameID int64,
) ([]sqlc.GetRegionsByGameRow, error) {
	regions, err := s.querier.GetRegionsByGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}

	s.log.Infow("got regions", "regions", regions)

	return regions, nil
}
