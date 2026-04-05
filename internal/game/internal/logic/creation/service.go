package creation

import (
	"fmt"
	"time"

	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/card"
	gamemetrics "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/metrics"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/mission"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/region"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	CreateGame(
		ctx kernelctx.UserContext,
		lobbyID int64,
		regions []string,
		players []player.Player,
	) (int64, error)
	CreateGameWithQuerier(
		ctx kernelctx.UserContext,
		querier db.Querier,
		regions []string,
		players []player.Player,
	) (int64, error)
}

type service struct {
	querier        db.Querier
	bus            eventbus.Publisher
	snapshotReader gameapi.SnapshotReader
	stateStore     gameapi.StateStore
	cardService    card.Service
	missionService mission.Service
	playerService  player.Service
	regionService  region.Service
	stateMetrics   *metrics.StateMetrics
	gameMetrics    *gamemetrics.GameMetrics
	gameTiming     *gamemetrics.GameTiming
}

var _ Service = (*service)(nil)

func NewService(
	querier db.Querier,
	bus eventbus.Publisher,
	snapshotReader gameapi.SnapshotReader,
	stateStore gameapi.StateStore,
	cardService card.Service,
	missionService mission.Service,
	playerService player.Service,
	regionService region.Service,
	stateMetrics *metrics.StateMetrics,
	gameMetrics *gamemetrics.GameMetrics,
	gameTiming *gamemetrics.GameTiming,
) Service {
	return &service{
		querier:        querier,
		bus:            bus,
		snapshotReader: snapshotReader,
		stateStore:     stateStore,
		playerService:  playerService,
		missionService: missionService,
		regionService:  regionService,
		cardService:    cardService,
		stateMetrics:   stateMetrics,
		gameMetrics:    gameMetrics,
		gameTiming:     gameTiming,
	}
}

func (s *service) CreateGame(
	userCtx kernelctx.UserContext,
	lobbyID int64,
	regions []string,
	players []player.Player,
) (int64, error) {
	gameID, err := dbutil.InTransaction(
		s.querier,
		userCtx,
		s.stateMetrics,
		func(qtx db.Querier) (int64, error) {
			return s.CreateGameWithQuerier(userCtx, qtx, regions, players)
		},
	)
	if err != nil {
		return -1, fmt.Errorf("failed to create game: %w", err)
	}

	s.gameMetrics.ActiveGames.Add(userCtx, 1)
	s.gameTiming.RecordStart(gameID)

	// Post-commit: read initial snapshot from the pool querier (committed data)
	// and store it in the StateStore for zero-query PlayerConnected cache.
	gameCtx := ctx.WithGameID(userCtx, gameID)

	publicSnap, privateSnaps, err := s.readInitialSnapshot(gameCtx)
	if err != nil {
		// Snapshot read failure is non-fatal: the game is already committed.
		// PlayerConnected will fall back to a fresh DB read if StateStore is empty.
		// Emit a lean event without snapshots.
		s.bus.Emit(
			userCtx,
			gameevt.NewGameCreated(gameID, lobbyID, time.Now(), len(players), nil, nil),
		)

		//nolint:nilerr // snapshot read failure is non-fatal (see comment above)
		return gameID, nil
	}

	s.stateStore.Store(gameID, &snapshot.CachedGameState{
		Turn:             0,
		PublicSnapshot:   publicSnap,
		PrivateSnapshots: privateSnaps,
	})

	s.bus.Emit(
		userCtx,
		gameevt.NewGameCreated(
			gameID,
			lobbyID,
			time.Now(),
			len(players),
			publicSnap,
			privateSnaps,
		),
	)

	return gameID, nil
}

// readInitialSnapshot reads the initial game state after transaction commit.
// Uses the pool querier (not a transactional one) since the data is committed.
func (s *service) readInitialSnapshot(
	gameCtx ctx.GameContext,
) (*snapshot.GameSnapshot, map[string]*snapshot.PlayerPrivate, error) {
	publicSnap, err := s.snapshotReader.GetPublicSnapshot(gameCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("reading public snapshot: %w", err)
	}

	privateSnaps, err := s.snapshotReader.GetAllPrivateSnapshots(gameCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("reading private snapshots: %w", err)
	}

	return publicSnap, privateSnaps, nil
}

func (s *service) CreateGameWithQuerier(
	userCtx kernelctx.UserContext,
	querier db.Querier,
	regions []string,
	players []player.Player,
) (int64, error) {
	return observe.Span(
		userCtx,
		"game.create",
		func(userCtx kernelctx.UserContext) (int64, error) {
			game, err := querier.InsertGame(userCtx)
			if err != nil {
				return -1, fmt.Errorf("failed to insert game: %w", err)
			}

			gameCtx := ctx.WithGameID(userCtx, game.ID)

			createdPlayers, err := s.playerService.CreatePlayers(
				gameCtx,
				querier,
				game.ID,
				players,
			)
			if err != nil {
				return -1, fmt.Errorf("failed to create players: %w", err)
			}

			if err = s.missionService.CreateMissions(
				gameCtx, querier, createdPlayers,
			); err != nil {
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

			if err := s.createPhase(gameCtx, querier, game); err != nil {
				return -1, fmt.Errorf("failed to create phase: %w", err)
			}

			return game.ID, nil
		},
	)
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

	if err := querier.SetGamePhase(ctx, sqlc.SetGamePhaseParams{
		ID:             game.ID,
		CurrentPhaseID: pgtype.Int8{Int64: phase.ID, Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	if _, err = querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          phase.ID,
		DeployableTroops: int64(3),
	}); err != nil {
		return fmt.Errorf("failed to create deploy phase: %w", err)
	}

	return nil
}
