package region

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/region/assignment"
	"go.uber.org/zap"
)

type Service interface {
	CreateRegions(
		ctx context.Context,
		q db.Querier,
		players []db.Player,
		regions []board.Region,
	) error
}

type ServiceImpl struct {
	log               *zap.SugaredLogger
	assignmentService assignment.Service
}

func NewRegionService(log *zap.SugaredLogger, assignmentService assignment.Service) *ServiceImpl {
	return &ServiceImpl{log: log, assignmentService: assignmentService}
}

func (s *ServiceImpl) CreateRegions(
	ctx context.Context,
	querier db.Querier,
	players []db.Player,
	regions []board.Region,
) error {
	s.log.Infow("creating regions", "players", players, "regions", regions)

	regionToPlayer := s.assignmentService.AssignRegionsToPlayers(players, regions)
	regionsParams := make([]db.InsertRegionsParams, 0, len(regionToPlayer))

	for region := range regionToPlayer {
		regionsParams = append(regionsParams, db.InsertRegionsParams{
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
