package game

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/data/db"
	sqlc "github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"github.com/tomfran/go-risk-it/internal/logic/region"
	"go.uber.org/zap"
)

type Service interface {
	CreateGameWithTx(ctx context.Context, board *board.Board, users []string) error
	CreateGame(ctx context.Context, querier db.Querier, board *board.Board, users []string) error
	GetGameState(ctx context.Context, gameID int64) (*sqlc.Game, error)
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

func (s *ServiceImpl) CreateGameWithTx(
	ctx context.Context,
	board *board.Board,
	users []string,
) error {
	err := s.querier.ExecuteInTransaction(ctx, func(qtx db.Querier) error {
		return s.CreateGame(ctx, qtx, board, users)
	})
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	return nil
}

func (s *ServiceImpl) CreateGame(
	ctx context.Context,
	querier db.Querier,
	board *board.Board,
	users []string,
) error {
	s.log.Infow("creating game", "board", board, "users", users)

	gameID, err := querier.InsertGame(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert game: %w", err)
	}

	s.log.Infow("inserted game", "id", gameID)

	players, err := s.playerService.CreatePlayers(ctx, querier, gameID, users)
	if err != nil {
		return fmt.Errorf("failed to create players: %w", err)
	}

	if err := s.regionService.CreateRegions(ctx, querier, players, board.Regions); err != nil {
		return fmt.Errorf("failed to create regions: %w", err)
	}

	s.log.Infow("successfully created game", "board", board, "users", users)

	return nil
}

func (s *ServiceImpl) GetGameState(
	ctx context.Context,
	gameID int64,
) (*sqlc.Game, error) {
	game, err := s.querier.GetGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return &game, nil
}
