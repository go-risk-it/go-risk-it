package region

import (
	"context"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/game/board"
	"github.com/tomfran/go-risk-it/internal/game/region/assignment"
	"go.uber.org/zap"
)

type Service struct {
	q                 *db.Queries
	log               *zap.SugaredLogger
	assignmentService *assignment.Service
}

func NewRegionService(queries *db.Queries, log *zap.SugaredLogger, assignmentService *assignment.Service) *Service {
	return &Service{q: queries, log: log, assignmentService: assignmentService}
}

func (s *Service) CreateRegions(players []db.Player, regions []board.Region) error {
	s.log.Infow("creating regions", "players", players, "regions", regions)
	regionToPlayer := s.assignmentService.AssignRegionsToPlayers(players, regions)
	var regionsParams []db.InsertRegionsParams
	for region := range regionToPlayer {
		regionsParams = append(regionsParams, db.InsertRegionsParams{
			PlayerID: regionToPlayer[region].ID,
		})
	}
	if _, err := s.q.InsertRegions(context.Background(), regionsParams); err != nil {
		s.log.Errorw("failed to insert regions", "error", err)
		return err
	}
	s.log.Infow("created regions", "players", players, "regions", regions)
	return nil
}
