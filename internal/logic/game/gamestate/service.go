package gamestate

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"go.uber.org/zap"
)

type Service interface {
	CreateGameWithTx(
		ctx context.Context,
		board *board.Board,
		players []request.Player,
	) (int64, error)
	CreateGameQ(
		ctx context.Context,
		querier db.Querier,
		board *board.Board,
		players []request.Player) (int64, error)
	GetGameState(ctx context.Context, gameID int64) (*sqlc.Game, error)
	GetGameStateQ(ctx context.Context, querier db.Querier, gameID int64) (*sqlc.Game, error)
	DecreaseDeployableTroopsQ(
		ctx context.Context,
		querier db.Querier,
		game *sqlc.Game,
		troops int64,
	) error
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
	players []request.Player,
) (int64, error) {
	gameID, err := s.querier.ExecuteInTransaction(ctx, func(qtx db.Querier) (interface{}, error) {
		return s.CreateGameQ(ctx, qtx, board, players)
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

func (s *ServiceImpl) CreateGameQ(
	ctx context.Context,
	querier db.Querier,
	board *board.Board,
	players []request.Player,
) (int64, error) {
	s.log.Debugw("creating game", "board", board, "players", players)

	game, err := querier.InsertGame(ctx, 3)
	if err != nil {
		return -1, fmt.Errorf("failed to insert game: %w", err)
	}

	s.log.Debugw("inserted game", "id", game)

	createdPlayers, err := s.playerService.CreatePlayers(ctx, querier, game.ID, players)
	if err != nil {
		return -1, fmt.Errorf("failed to create players: %w", err)
	}

	err = s.regionService.CreateRegions(ctx, querier, createdPlayers, board.Regions)
	if err != nil {
		return -1, fmt.Errorf("failed to create regions: %w", err)
	}

	s.log.Debugw("successfully created game", "board", board, "players", players)

	return game.ID, nil
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

func (s *ServiceImpl) DecreaseDeployableTroopsQ(
	ctx context.Context,
	querier db.Querier,
	game *sqlc.Game,
	troops int64,
) error {
	s.log.Infow("decreasing deployable troops", "gameID", game.ID, "troops", troops)

	if game.DeployableTroops < troops {
		return fmt.Errorf("player does not have enough troops to deploy")
	}

	err := querier.DecreaseDeployableTroops(ctx, sqlc.DecreaseDeployableTroopsParams{
		ID:               game.ID,
		DeployableTroops: troops,
	})
	if err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	s.log.Infow("decreased deployable troops", "gameID", game.ID, "troops", troops)

	return nil
}
