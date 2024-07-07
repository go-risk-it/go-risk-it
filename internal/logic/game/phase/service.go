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
	CreateNewPhaseQ(
		ctx ctx.UserContext,
		querier db.Querier,
		gameID, turn int64,
		phaseType sqlc.PhaseType,
	) (int64, error)
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

	targetPhaseType, err := s.phaseWalker.WalkToTargetPhase(ctx, querier, gameState.Phase)
	if err != nil {
		return fmt.Errorf("failed to walk to target phase: %w", err)
	}

	if targetPhaseType == gameState.Phase {
		ctx.Log().Infow("no need to advance phase")

		return nil
	}

	ctx.Log().Infow(
		"advancing phase",
		"from",
		gameState.Phase,
		"to",
		targetPhaseType,
	)

	targetPhaseID, err := s.CreateNewPhaseQ(
		ctx,
		querier,
		gameState.ID,
		gameState.Turn,
		targetPhaseType,
	)
	if err != nil {
		return fmt.Errorf("failed to create new phase: %w", err)
	}

	err = s.SetGamePhaseQ(ctx, querier, ctx.GameID(), targetPhaseID)
	if err != nil {
		return fmt.Errorf("failed to set game phase: %w", err)
	}

	return nil
}

func (s *ServiceImpl) CreateNewPhaseQ(
	ctx ctx.UserContext,
	querier db.Querier,
	gameID, turn int64,
	phaseType sqlc.PhaseType,
) (int64, error) {
	var (
		phaseID int64
		err     error
	)

	switch phaseType {
	case sqlc.PhaseTypeDEPLOY:
		phaseID, err = s.createDeployPhaseQ(ctx, querier, gameID, turn)
	case sqlc.PhaseTypeATTACK:
		phaseID, err = s.createAttackPhaseQ(ctx, querier, gameID, turn)
	default:
		err = fmt.Errorf("unsupported phase type: %v", phaseType)
	}

	if err != nil {
		return -1, fmt.Errorf("failed to create deploy phase: %w", err)
	}

	return phaseID, nil
}

func (s *ServiceImpl) createDeployPhaseQ(
	ctx ctx.UserContext,
	querier db.Querier,
	gameID int64,
	turn int64,
) (int64, error) {
	ctx.Log().Infow("creating deploy phase", "gameID", gameID, "turn", turn)

	phase, err := s.insertPhaseQ(ctx, querier, gameID, sqlc.PhaseTypeDEPLOY, turn)
	if err != nil {
		return -1, fmt.Errorf("failed to create deploy phase: %w", err)
	}

	deployableTroops := int64(3)

	deployPhase, err := querier.InsertDeployPhase(ctx, sqlc.InsertDeployPhaseParams{
		PhaseID:          phase.ID,
		DeployableTroops: deployableTroops,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to create deploy phase: %w", err)
	}

	ctx.Log().Infow("deploy phase created", "phase", deployPhase)

	return deployPhase.PhaseID, nil
}

func (s *ServiceImpl) createAttackPhaseQ(
	ctx ctx.UserContext,
	querier db.Querier,
	gameID int64,
	turn int64,
) (int64, error) {
	ctx.Log().Infow("creating attack phase", "gameID", gameID, "turn", turn)

	phase, err := s.insertPhaseQ(ctx, querier, gameID, sqlc.PhaseTypeATTACK, turn)
	if err != nil {
		return -1, fmt.Errorf("failed to create attack phase: %w", err)
	}

	ctx.Log().Infow("attack phase created", "phase", phase)

	return phase.ID, nil
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
