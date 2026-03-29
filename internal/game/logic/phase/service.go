package phase

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	InsertPhase(
		ctx ctx.GameContext,
		querier db.Querier,
		phaseType sqlc.GamePhaseType,
	) (*sqlc.GamePhase, error)
}

type service struct {
	gameService   state.Service
	playerService player.Service
}

var _ Service = (*service)(nil)

func NewService(gameService state.Service, playerService player.Service) Service {
	return &service{
		gameService:   gameService,
		playerService: playerService,
	}
}

func (s *service) InsertPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	phaseType sqlc.GamePhaseType,
) (*sqlc.GamePhase, error) {
	slog.InfoContext(ctx, "checking if phase needs to be advanced")

	gameState, err := s.gameService.GetGameStateWithQuerier(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	if phaseType == gameState.Phase {
		return nil, fmt.Errorf("game already in desired phase: %v", phaseType)
	}

	slog.InfoContext(ctx,
		"inserting phase",
		"phase",
		phaseType,
	)

	currentPhase, err := querier.GetCurrentPhase(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("failed to get current phase: %w", err)
	}

	turn, err := s.getNextTurn(ctx, querier, gameState, currentPhase)
	if err != nil {
		return nil, fmt.Errorf("failed to get next turn: %w", err)
	}

	phase, err := s.insertPhase(ctx, querier, ctx.GameID(), phaseType, turn)
	if err != nil {
		return nil, fmt.Errorf("failed to create new phase: %w", err)
	}

	err = s.setGamePhase(ctx, querier, ctx.GameID(), phase.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to set game phase: %w", err)
	}

	return phase, nil
}

func (s *service) getNextTurn(
	ctx ctx.GameContext,
	querier db.Querier,
	gameState *state.Game,
	currentPhase sqlc.GamePhaseType,
) (int64, error) {
	turn := gameState.Turn

	if currentPhase == sqlc.GamePhaseTypeREINFORCE {
		playersState, err := s.playerService.GetPlayersStateWithQuerier(ctx, querier)
		if err != nil {
			return 0, fmt.Errorf("failed to get players state: %w", err)
		}

		turn++

		players := int64(len(playersState))
		for playersState[turn%players].RegionCount == 0 {
			turn++
		}
	}

	return turn, nil
}

func (s *service) insertPhase(
	ctx ctx.UserContext,
	querier db.Querier,
	gameID int64,
	phaseType sqlc.GamePhaseType,
	turn int64,
) (*sqlc.GamePhase, error) {
	slog.InfoContext(ctx, "creating phase", "gameID", gameID, "turn", turn)

	phase, err := querier.InsertPhase(ctx, sqlc.InsertPhaseParams{
		GameID: gameID,
		Type:   phaseType,
		Turn:   turn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create phase: %w", err)
	}

	slog.InfoContext(ctx, "phase created", "phase", phase)

	return &phase, nil
}

func (s *service) setGamePhase(
	ctx ctx.UserContext,
	querier db.Querier,
	gameID int64,
	phaseID int64,
) error {
	slog.DebugContext(ctx, "setting phase", "phase", phaseID)

	err := querier.SetGamePhase(ctx, sqlc.SetGamePhaseParams{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: phaseID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to set phase: %w", err)
	}

	slog.DebugContext(ctx, "phase set", "phase", phaseID)

	return nil
}
