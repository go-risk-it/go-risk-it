package phase

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	InsertPhaseQ(
		ctx ctx.GameContext,
		querier db.Querier,
		phaseType sqlc.PhaseType,
	) (*sqlc.Phase, error)
}

type ServiceImpl struct {
	gameService state.Service
}

var _ Service = &ServiceImpl{}

func NewService(gameService state.Service) *ServiceImpl {
	return &ServiceImpl{
		gameService: gameService,
	}
}

func (s *ServiceImpl) InsertPhaseQ(
	ctx ctx.GameContext,
	querier db.Querier,
	phaseType sqlc.PhaseType,
) (*sqlc.Phase, error) {
	ctx.Log().Infow("checking if phase needs to be advanced")

	gameState, err := s.gameService.GetGameStateQ(ctx, querier)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	if phaseType == gameState.Phase {
		return nil, fmt.Errorf("game already in desired phase: %v", phaseType)
	}

	ctx.Log().Infow(
		"inserting phase",
		"phase",
		phaseType,
	)

	turn := gameState.Turn
	if phaseType == sqlc.PhaseTypeCARDS {
		turn++
	}

	phase, err := s.insertPhaseQ(ctx, querier, ctx.GameID(), phaseType, turn)
	if err != nil {
		return nil, fmt.Errorf("failed to create new phase: %w", err)
	}

	err = s.setGamePhaseQ(ctx, querier, ctx.GameID(), phase.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to set game phase: %w", err)
	}

	return phase, nil
}

func (s *ServiceImpl) insertPhaseQ(
	ctx ctx.UserContext,
	querier db.Querier,
	gameID int64,
	phaseType sqlc.PhaseType,
	turn int64,
) (*sqlc.Phase, error) {
	ctx.Log().Infow("creating phase", "gameID", gameID, "turn", turn)

	phase, err := querier.InsertPhase(ctx, sqlc.InsertPhaseParams{
		GameID: gameID,
		Type:   phaseType,
		Turn:   turn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create phase: %w", err)
	}

	ctx.Log().Infow("phase created", "phase", phase)

	return &phase, nil
}

func (s *ServiceImpl) setGamePhaseQ(
	ctx ctx.UserContext,
	querier db.Querier,
	gameID int64,
	phaseID int64,
) error {
	ctx.Log().Infow("setting phase", "phase", phaseID)

	err := querier.SetGamePhase(ctx, sqlc.SetGamePhaseParams{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: phaseID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to set phase: %w", err)
	}

	ctx.Log().Infow("phase set", "phase", phaseID)

	return nil
}
