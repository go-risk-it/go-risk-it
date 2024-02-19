package region

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/data/db"
	sqlc "github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/region/assignment"
	"go.uber.org/zap"
)

type Service interface {
	CreateRegions(
		ctx context.Context,
		querier db.Querier,
		players []sqlc.Player,
		regions []board.Region,
	) error
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

	regionToPlayer := s.assignmentService.AssignRegionsToPlayers(players, regions)
	regionsParams := make([]sqlc.InsertRegionsParams, 0, len(regionToPlayer))

	for region := range regionToPlayer {
		regionsParams = append(regionsParams, sqlc.InsertRegionsParams{
			PlayerID: regionToPlayer[region].ID,
			Troops:   3,
		})
	}

	if _, err := querier.InsertRegions(ctx, regionsParams); err != nil {
		return fmt.Errorf("failed to insert regions: %w", err)
	}

	s.log.Infow("created regions", "players", players, "regions", regions)

	return nil
}
