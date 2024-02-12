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
	CreateGame(ctx context.Context, board *board.Board, users []string) error
	GetGameState(ctx context.Context, gameID int64) (*db.Game, error)
}

type ServiceImpl struct {
	log           *zap.SugaredLogger
	querier       db.Querier
	playerService player.Service
	regionService region.Service
}

func NewGameService(
	logger *zap.SugaredLogger,
	querier db.Querier,
	playerService player.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		log:           logger,
		querier:       querier,
		playerService: playerService,
		regionService: regionService,
	}
}

func (s *ServiceImpl) CreateGame(
	ctx context.Context,
	board *board.Board,
	users []string,
) error {
	s.log.Infow("creating logic", "board", board, "users", users)

	gameID, err := s.querier.InsertGame(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert game: %w", err)
	}

	players, err := s.playerService.CreatePlayers(ctx, gameID, users)
	if err != nil {
		return fmt.Errorf("failed to create players: %w", err)
	}

	if err := s.regionService.CreateRegions(ctx, s.querier, players, board.Regions); err != nil {
		return fmt.Errorf("failed to create regions: %w", err)
	}

	s.log.Infow("created logic", "board", board, "users", users)

	return nil
}

func (s *ServiceImpl) GetGameState(
	ctx context.Context,
	gameID int64,
) (*db.Game, error) {
	game, err := s.querier.GetGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return &game, nil
}
