package game

import (
	"context"

	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/game/board"
	"github.com/tomfran/go-risk-it/internal/game/player"
	"github.com/tomfran/go-risk-it/internal/game/region"
	"go.uber.org/zap"
)

type Service interface {
	CreateGame(ctx context.Context, q db.Querier, board *board.Board, users []string) error
}

type ServiceImpl struct {
	log           *zap.SugaredLogger
	playerService player.Service
	regionService region.Service
}

func NewGameService(logger *zap.SugaredLogger, playerService player.Service, regionService region.Service) *ServiceImpl {
	return &ServiceImpl{log: logger, playerService: playerService, regionService: regionService}
}

func (s *ServiceImpl) CreateGame(ctx context.Context, q db.Querier, board *board.Board, users []string) error {
	s.log.Infow("creating game", "board", board, "users", users)
	gameId, err := q.InsertGame(ctx)
	if err != nil {
		return err
	}

	players, err := s.playerService.CreatePlayers(ctx, q, gameId, users)
	if err != nil {
		return err
	}

	if err := s.regionService.CreateRegions(ctx, q, players, board.Regions); err != nil {
		return err
	}
	s.log.Infow("created game", "board", board, "users", users)

	return nil
}
