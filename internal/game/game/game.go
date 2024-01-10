package game

import (
	"context"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/game/board"
	"github.com/tomfran/go-risk-it/internal/game/player"
	"github.com/tomfran/go-risk-it/internal/game/region"
	"github.com/tomfran/go-risk-it/internal/game/region/assignment"
	"go.uber.org/zap"
)

type Service struct {
	q             *db.Queries
	log           *zap.SugaredLogger
	playerService *player.Service
	regionService *region.Service
}

func NewGameService(queries *db.Queries, logger *zap.SugaredLogger, playerService *player.Service, assignmentService *assignment.Service, regionService *region.Service) *Service {
	return &Service{q: queries, log: logger, playerService: playerService, regionService: regionService}
}

func (s *Service) CreateGame(board *board.Board, users []string) error {
	s.log.Infow("creating game", "board", board, "users", users)
	gameId, err := s.q.InsertGame(context.Background())
	if err != nil {
		return err
	}
	players, err := s.playerService.CreatePlayers(gameId, users)
	if err != nil {
		return err
	}
	if err := s.regionService.CreateRegions(players, board.Regions); err != nil {
		return err
	}
	s.log.Infow("created game", "board", board, "users", users)

	return nil
}
