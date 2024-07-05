package phase

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/deploy"
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
	attackService attack.Service
	deployService deploy.Service
	gameService   state.Service
}

var _ Service = &ServiceImpl{}

func NewService(
	attackService attack.Service,
	deployService deploy.Service,
	gameService state.Service,
) *ServiceImpl {
	return &ServiceImpl{
		attackService: attackService,
		deployService: deployService,
		gameService:   gameService,
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

	targetPhaseType, err := s.walkToTargetPhase(ctx, querier, gameState)
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

func (s *ServiceImpl) walkToTargetPhase(
	ctx ctx.MoveContext,
	querier db.Querier,
	gameState *state.Game,
) (sqlc.PhaseType, error) {
	targetPhase := gameState.CurrentPhase

	mustAdvance := true
	for mustAdvance {
		mustAdvance = false

		switch targetPhase {
		case sqlc.PhaseTypeDEPLOY:
			deployableTroops, err := s.deployService.GetDeployableTroops(ctx, querier)
			if err != nil {
				return targetPhase, fmt.Errorf("failed to get deployable troops: %w", err)
			}

			if deployableTroops == 0 {
				ctx.Log().Infow(
					"deploy must advance",
					"phase",
					gameState.CurrentPhase,
				)

				targetPhase = sqlc.PhaseTypeATTACK
				mustAdvance = true
			}
		case sqlc.PhaseTypeATTACK:
			targetPhase, err := s.getTargetPhaseForAttack(ctx, querier)
			if err != nil {
				return targetPhase, fmt.Errorf("failed to get target phase for attack: %w", err)
			}

			if targetPhase != sqlc.PhaseTypeATTACK {
				mustAdvance = true
			}
		case sqlc.PhaseTypeCONQUER:
		case sqlc.PhaseTypeREINFORCE:
		case sqlc.PhaseTypeCARDS:
		}
	}

	return targetPhase, nil
}

func (s *ServiceImpl) getTargetPhaseForAttack(
	ctx ctx.MoveContext,
	querier db.Querier,
) (sqlc.PhaseType, error) {
	targetPhase := sqlc.PhaseTypeATTACK

	hasConquered, err := s.attackService.HasConqueredQ(ctx, querier)
	if err != nil {
		return targetPhase, fmt.Errorf("failed to check if attack has conquered: %w", err)
	}

	if hasConquered {
		ctx.Log().Infow("must advance phase to CONQUER")

		return sqlc.PhaseTypeCONQUER, nil
	}

	canContinueAttacking, err := s.attackService.CanContinueAttackingQ(ctx, querier)
	if err != nil {
		return targetPhase, fmt.Errorf("failed to check if attack can continue: %w", err)
	}

	if !canContinueAttacking {
		ctx.Log().Infow("cannot continue attacking, must advance phase to REINFORCE")

		return sqlc.PhaseTypeREINFORCE, nil
	}

	return targetPhase, nil
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
