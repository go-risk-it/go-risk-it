package phase

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase/walker"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	AdvanceQ(ctx ctx.MoveContext, querier db.Querier) error
	SetGamePhaseQ(ctx ctx.UserContext, querier db.Querier, gameID int64, phaseID int64) error
	CreateDeployPhaseQ(
		ctx ctx.UserContext,
		querier db.Querier,
		gameID int64,
		turn int64,
		deployableTroops int64,
	) (*sqlc.DeployPhase, error)
}

type ServiceImpl struct {
	gameService state.Service
	phaseWalker walker.Service
}

var _ Service = &ServiceImpl{}

func NewService(gameService state.Service, phaseWalker walker.Service) *ServiceImpl {
	return &ServiceImpl{
		gameService: gameService,
		phaseWalker: phaseWalker,
	}
}

func (s *ServiceImpl) SetGamePhaseQ(
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

func (s *ServiceImpl) AdvanceQ(ctx ctx.MoveContext, querier db.Querier) error {
	ctx.Log().Infow("checking if phase needs to be advanced")

	gameState, err := s.gameService.GetGameStateQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	ctx.Log().Infow("walking to target phase", "from", gameState.CurrentPhase)

	targetPhaseType, err := s.phaseWalker.WalkToTargetPhase(ctx, querier, gameState.CurrentPhase)
	if err != nil {
		return fmt.Errorf("failed to walk to target phase: %w", err)
	}

	if targetPhaseType == gameState.CurrentPhase {
		ctx.Log().Infow("no need to advance phase")

		return nil
	}

	ctx.Log().Infow(
		"advancing phase",
		"from",
		gameState.CurrentPhase,
		"to",
		targetPhaseType,
	)

	// TODO: Create phase
	targetPhase := sqlc.Phase{
		ID:   0,
		Type: targetPhaseType,
		Turn: gameState.CurrentTurn,
	}

	err = s.SetGamePhaseQ(ctx, querier, ctx.GameID(), targetPhase.ID)
	if err != nil {
		return fmt.Errorf("failed to set game phase: %w", err)
	}

	return nil
}

func (s *ServiceImpl) CreateDeployPhaseQ(
	ctx ctx.UserContext,
	querier db.Querier,
	gameID int64,
	turn int64,
	deployableTroops int64,
) (*sqlc.DeployPhase, error) {
	ctx.Log().Infow("creating deploy phase", "gameID", gameID, "turn", turn)

	phase, err := s.createPhaseQ(ctx, querier, gameID, sqlc.PhaseTypeDEPLOY, turn)
	if err != nil {
		return nil, fmt.Errorf("failed to create deploy phase: %w", err)
	}

	deployPhase, err := querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          phase.ID,
		DeployableTroops: deployableTroops,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create deploy phase: %w", err)
	}

	ctx.Log().Infow("deploy phase created", "phase", deployPhase)

	return &deployPhase, nil
}

func (s *ServiceImpl) createPhaseQ(
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
