package creation

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/card"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/jackc/pgx/v5/pgtype"
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
	querier        db.Querier
	cardService    card.Service
	missionService mission.Service
	playerService  player.Service
	regionService  region.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
	cardService card.Service,
	missionService mission.Service,
	playerService player.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		querier:        querier,
		playerService:  playerService,
		missionService: missionService,
		regionService:  regionService,
		cardService:    cardService,
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

	ctx := ctx.WithGameID(cont, game.ID)

	createdPlayers, err := s.playerService.CreatePlayersQ(ctx, querier, game.ID, players)
	if err != nil {
		return -1, fmt.Errorf("failed to create players: %w", err)
	}

	if err = s.missionService.CreateMissionsQ(ctx, querier, createdPlayers); err != nil {
		return -1, fmt.Errorf("failed to create missions: %w", err)
	}

	if err = s.regionService.CreateRegionsQ(ctx, querier, createdPlayers, regions); err != nil {
		return -1, fmt.Errorf("failed to create regions: %w", err)
	}

	if err = s.cardService.CreateCardsQ(ctx, querier); err != nil {
		return -1, fmt.Errorf("failed to create cards: %w", err)
	}

	ctx.Log().Debugw("creating initial phase", "gameID", game.ID)

	if err := s.createPhase(ctx, querier, game); err != nil {
		return -1, fmt.Errorf("failed to create phase: %w", err)
	}

	ctx.Log().
		Infow("successfully created game", "regions", len(regions), "players", len(players))

	return game.ID, nil
}

func (s *ServiceImpl) createPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	game sqlc.Game,
) error {
	phase, err := querier.InsertPhase(ctx, sqlc.InsertPhaseParams{
		GameID: game.ID,
		Type:   sqlc.PhaseTypeDEPLOY,
		Turn:   0,
	})
	if err != nil {
		return fmt.Errorf("failed to create initial phase: %w", err)
	}

	ctx.Log().
		Infow("updating game phase", "gameID", game.ID, "phaseID", phase.ID)

	if err := querier.SetGamePhase(ctx, sqlc.SetGamePhaseParams{
		ID:             game.ID,
		CurrentPhaseID: pgtype.Int8{Int64: phase.ID, Valid: true},
	}); err != nil {
		ctx.Log().Warnw("failed to update game phase", "err", err)

		return fmt.Errorf("failed to update game phase: %w", err)
	}

	ctx.Log().Infow("updated phase, creating deploy phase")

	if _, err = querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          phase.ID,
		DeployableTroops: int64(3),
	}); err != nil {
		return fmt.Errorf("failed to create deploy phase: %w", err)
	}

	return nil
}
