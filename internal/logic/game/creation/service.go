package creation

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
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
}

type ServiceImpl struct {
	querier       db.Querier
	cardsService  cards.Service
	playerService player.Service
	regionService region.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
	cardsService cards.Service,
	playerService player.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		querier:       querier,
		cardsService:  cardsService,
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
	cont ctx.UserContext,
	querier db.Querier,
	regions []string,
	players []request.Player,
) (int64, error) {
	cont.Log().Infow("creating game", "regions", len(regions), "players", len(players))

	game, err := querier.InsertGame(cont)
	if err != nil {
		return -1, fmt.Errorf("failed to insert game: %w", err)
	}

	cont.Log().Debugw("inserted game, creating initial phase", "gameID", game.ID)

	gameContext := ctx.WithGameID(cont, game.ID)
	moveContext := ctx.NewMoveContext(cont, gameContext)

	err = s.cardsService.AdvanceQ(
		moveContext,
		querier,
		sqlc.PhaseTypeDEPLOY,
		cards.Move{},
	)
	if err != nil {
		return -1, fmt.Errorf("failed to create initial phase: %w", err)
	}

	createdPlayers, err := s.playerService.CreatePlayers(cont, querier, game.ID, players)
	if err != nil {
		return -1, fmt.Errorf("failed to create players: %w", err)
	}

	err = s.regionService.CreateRegions(cont, querier, createdPlayers, regions)
	if err != nil {
		return -1, fmt.Errorf("failed to create regions: %w", err)
	}

	cont.Log().Infow("successfully created game", "regions", len(regions), "players", len(players))

	return game.ID, nil
}
