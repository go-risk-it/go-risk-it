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
	CreateGameWithTx(ctx context.Context, board *board.Board, users []string) (int64, error)
	CreateGame(
		ctx context.Context,
		querier db.Querier,
		board *board.Board,
		users []string) (int64, error)
	GetGameState(ctx context.Context, gameID int64) (*sqlc.Game, error)
	GetGameStateQ(ctx context.Context, querier db.Querier, gameID int64) (*sqlc.Game, error)
	SetGamePhaseQ(ctx context.Context, querier db.Querier, gameID int64, phase sqlc.Phase) error
}

type ServiceImpl struct {
	log           *zap.SugaredLogger
	querier       db.Querier
	playerService player.Service
	regionService region.Service
}

func NewService(
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
) (int64, error) {
	gameID, err := s.querier.ExecuteInTransaction(ctx, func(qtx db.Querier) (interface{}, error) {
		return s.CreateGame(ctx, qtx, board, users)
	})
	if err != nil {
		return -1, fmt.Errorf("failed to create game: %w", err)
	}

	gameIDInt, ok := gameID.(int64)
	if !ok {
		return -1, fmt.Errorf("failed to convert gameID to int64: %w", err)
	}

	return gameIDInt, nil
}

func (s *ServiceImpl) CreateGame(
	ctx context.Context,
	querier db.Querier,
	board *board.Board,
	users []string,
) (int64, error) {
	s.log.Infow("creating game", "board", board, "users", users)

	gameID, err := querier.InsertGame(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to insert game: %w", err)
	}

	s.log.Infow("inserted game", "id", gameID)

	players, err := s.playerService.CreatePlayers(ctx, querier, gameID, users)
	if err != nil {
		return -1, fmt.Errorf("failed to create players: %w", err)
	}

	if err := s.regionService.CreateRegions(ctx, querier, players, board.Regions); err != nil {
		return -1, fmt.Errorf("failed to create regions: %w", err)
	}

	s.log.Infow("successfully created game", "board", board, "users", users)

	return gameID, nil
}

func (s *ServiceImpl) GetGameState(
	ctx context.Context,
	gameID int64,
) (*sqlc.Game, error) {
	return s.GetGameStateQ(ctx, s.querier, gameID)
}

func (s *ServiceImpl) GetGameStateQ(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
) (*sqlc.Game, error) {
	game, err := querier.GetGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return &game, nil
}

func (s *ServiceImpl) SetGamePhaseQ(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	phase sqlc.Phase,
) error {
	s.log.Infow("setting phase", "gameID", gameID, "phase", phase)

	err := querier.SetGamePhase(ctx, sqlc.SetGamePhaseParams{
		Phase: phase,
		ID:    gameID,
	})
	if err != nil {
		return fmt.Errorf("failed to set phase: %w", err)
	}

	s.log.Infow("phase set", "gameID", gameID, "phase", phase)

	return nil
}
