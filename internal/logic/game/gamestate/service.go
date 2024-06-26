package gamestate

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
)

type Service interface {
	CreateGameWithTx(
		ctx ctx.UserContext,
		regions []string,
		players []request.Player,
	) (int64, error)
	CreateGameQ(
		ctx ctx.UserContext,
		querier db.Querier,
		regions []string,
		players []request.Player,
	) (int64, error)
	GetGameState(ctx ctx.GameContext) (*sqlc.Game, error)
	GetGameStateQ(ctx ctx.GameContext, querier db.Querier) (*sqlc.Game, error)
}

type ServiceImpl struct {
	querier       db.Querier
	playerService player.Service
	regionService region.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
	playerService player.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		querier:       querier,
		playerService: playerService,
		regionService: regionService,
	}
}

func (s *ServiceImpl) CreateGameWithTx(
	ctx ctx.UserContext,
	regions []string,
	players []request.Player,
) (int64, error) {
	gameID, err := s.querier.ExecuteInTransaction(ctx, func(qtx db.Querier) (interface{}, error) {
		return s.CreateGameQ(ctx, qtx, regions, players)
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
	ctx ctx.UserContext,
	querier db.Querier,
	regions []string,
	players []request.Player,
) (int64, error) {
	ctx.Log().Debugw("creating game", "regions", len(regions), "players", len(players))

	game, err := querier.InsertGame(ctx, 3)
	if err != nil {
		return -1, fmt.Errorf("failed to insert game: %w", err)
	}

	ctx.Log().Debugw("inserted game", "id", game)

	createdPlayers, err := s.playerService.CreatePlayers(ctx, querier, game.ID, players)
	if err != nil {
		return -1, fmt.Errorf("failed to create players: %w", err)
	}

	err = s.regionService.CreateRegions(ctx, querier, createdPlayers, regions)
	if err != nil {
		return -1, fmt.Errorf("failed to create regions: %w", err)
	}

	ctx.Log().Debugw("successfully created game", "regions", len(regions), "players", len(players))

	return game.ID, nil
}

func (s *ServiceImpl) GetGameState(ctx ctx.GameContext) (*sqlc.Game, error) {
	return s.GetGameStateQ(ctx, s.querier)
}

func (s *ServiceImpl) GetGameStateQ(ctx ctx.GameContext, querier db.Querier) (*sqlc.Game, error) {
	game, err := querier.GetGame(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return &game, nil
}
