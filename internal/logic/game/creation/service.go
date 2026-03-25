package creation

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/card"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/timing"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateGame(
		ctx ctx.UserContext,
		regions []string,
		players []player.Player,
	) (int64, error)
	CreateGameWithQuerier(
		ctx ctx.UserContext,
		querier db.Querier,
		regions []string,
		players []player.Player,
	) (int64, error)
}

type service struct {
	querier        db.Querier
	cardService    card.Service
	missionService mission.Service
	playerService  player.Service
	regionService  region.Service
	metrics        *metrics.Metrics
	gameTiming     *timing.GameTiming
}

var _ Service = (*service)(nil)

func NewService(
	querier db.Querier,
	cardService card.Service,
	missionService mission.Service,
	playerService player.Service,
	regionService region.Service,
	metrics *metrics.Metrics,
	gameTiming *timing.GameTiming,
) Service {
	return &service{
		querier:        querier,
		playerService:  playerService,
		missionService: missionService,
		regionService:  regionService,
		cardService:    cardService,
		metrics:        metrics,
		gameTiming:     gameTiming,
	}
}

func (s *service) CreateGame(
	ctx ctx.UserContext,
	regions []string,
	players []player.Player,
) (int64, error) {
	gameID, err := dbutil.InTransaction(
		s.querier,
		ctx,
		s.metrics,
		func(qtx db.Querier) (int64, error) {
			return s.CreateGameWithQuerier(ctx, qtx, regions, players)
		},
	)
	if err != nil {
		return -1, fmt.Errorf("failed to create game: %w", err)
	}

	s.metrics.GamesCreated.Add(ctx, 1)
	s.metrics.ActiveGames.Add(ctx, 1)
	s.gameTiming.RecordStart(gameID)

	return gameID, nil
}

func (s *service) CreateGameWithQuerier(
	userCtx ctx.UserContext,
	querier db.Querier,
	regions []string,
	players []player.Player,
) (int64, error) {
	slog.InfoContext(userCtx, "creating game", "regions", len(regions), "players", len(players))

	game, err := querier.InsertGame(userCtx)
	if err != nil {
		return -1, fmt.Errorf("failed to insert game: %w", err)
	}

	gameCtx := ctx.WithGameID(userCtx, game.ID)

	createdPlayers, err := s.playerService.CreatePlayers(gameCtx, querier, game.ID, players)
	if err != nil {
		return -1, fmt.Errorf("failed to create players: %w", err)
	}

	if err = s.missionService.CreateMissions(gameCtx, querier, createdPlayers); err != nil {
		return -1, fmt.Errorf("failed to create missions: %w", err)
	}

	if err = s.regionService.CreateRegions(
		gameCtx, querier, createdPlayers, regions,
	); err != nil {
		return -1, fmt.Errorf("failed to create regions: %w", err)
	}

	if err = s.cardService.CreateCards(gameCtx, querier); err != nil {
		return -1, fmt.Errorf("failed to create cards: %w", err)
	}

	slog.DebugContext(gameCtx, "creating initial phase", "gameID", game.ID)

	if err := s.createPhase(gameCtx, querier, game); err != nil {
		return -1, fmt.Errorf("failed to create phase: %w", err)
	}

	slog.InfoContext(gameCtx, "successfully created game",
		"regions", len(regions), "players", len(players))

	return game.ID, nil
}

func (s *service) createPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	game sqlc.GameGame,
) error {
	phase, err := querier.InsertPhase(ctx, sqlc.InsertPhaseParams{
		GameID: game.ID,
		Type:   sqlc.GamePhaseTypeDEPLOY,
		Turn:   0,
	})
	if err != nil {
		return fmt.Errorf("failed to create initial phase: %w", err)
	}

	slog.InfoContext(ctx, "updating game phase", "gameID", game.ID, "phaseID", phase.ID)

	if err := querier.SetGamePhase(ctx, sqlc.SetGamePhaseParams{
		ID:             game.ID,
		CurrentPhaseID: pgtype.Int8{Int64: phase.ID, Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	slog.InfoContext(ctx, "updated phase, creating deploy phase")

	if _, err = querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          phase.ID,
		DeployableTroops: int64(3),
	}); err != nil {
		return fmt.Errorf("failed to create deploy phase: %w", err)
	}

	return nil
}
