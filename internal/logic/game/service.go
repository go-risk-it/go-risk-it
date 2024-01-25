package game

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"github.com/tomfran/go-risk-it/internal/logic/region"
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

func NewGameService(
	logger *zap.SugaredLogger,
	playerService player.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{log: logger, playerService: playerService, regionService: regionService}
}

func (s *ServiceImpl) CreateGame(
	ctx context.Context,
	querier db.Querier,
	board *board.Board,
	users []string,
) error {
	s.log.Infow("creating logic", "board", board, "users", users)

	gameID, err := querier.InsertGame(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert game: %w", err)
	}

	players, err := s.playerService.CreatePlayers(ctx, querier, gameID, users)
	if err != nil {
		return fmt.Errorf("failed to create players: %w", err)
	}

	if err := s.regionService.CreateRegions(ctx, querier, players, board.Regions); err != nil {
		return fmt.Errorf("failed to create regions: %w", err)
	}

	s.log.Infow("created logic", "board", board, "users", users)

	return nil
}
