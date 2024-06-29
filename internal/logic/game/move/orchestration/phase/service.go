package phase

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/attack"
)

type Service interface {
	AdvanceQ(ctx ctx.MoveContext, querier db.Querier) error
	SetGamePhaseQ(ctx ctx.MoveContext, querier db.Querier, phase sqlc.Phase) error
}

type ServiceImpl struct {
	attackService attack.Service
	gameService   gamestate.Service
}

var _ Service = &ServiceImpl{}

func NewService(attackService attack.Service, gameService gamestate.Service) *ServiceImpl {
	return &ServiceImpl{
		attackService: attackService,
		gameService:   gameService,
	}
}

func (s *ServiceImpl) SetGamePhaseQ(
	ctx ctx.MoveContext,
	querier db.Querier,
	phase sqlc.Phase,
) error {
	ctx.Log().Infow("setting phase", "phase", phase)

	err := querier.SetGamePhase(ctx, sqlc.SetGamePhaseParams{
		Phase: phase,
		ID:    ctx.GameID(),
	})
	if err != nil {
		return fmt.Errorf("failed to set phase: %w", err)
	}

	ctx.Log().Infow("phase set", "phase", phase)

	return nil
}

func (s *ServiceImpl) AdvanceQ(ctx ctx.MoveContext, querier db.Querier) error {
	ctx.Log().Infow("checking if phase needs to be advanced")

	gameState, err := s.gameService.GetGameStateQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	ctx.Log().Infow("walking to target phase", "from", gameState.Phase)

	targetPhase, err := s.walkToTargetPhase(ctx, querier, gameState)
	if err != nil {
		return fmt.Errorf("failed to walk to target phase: %w", err)
	}

	if targetPhase == gameState.Phase {
		ctx.Log().Infow("no need to advance phase")

		return nil
	}

	ctx.Log().Infow(
		"advancing phase",
		"from",
		gameState.Phase,
		"to",
		targetPhase,
	)

	err = s.SetGamePhaseQ(ctx, querier, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to set game phase: %w", err)
	}

	return nil
}

func (s *ServiceImpl) walkToTargetPhase(
	ctx ctx.MoveContext,
	querier db.Querier,
	gameState *sqlc.Game,
) (sqlc.Phase, error) {
	targetPhase := gameState.Phase

	mustAdvance := true
	for mustAdvance {
		mustAdvance = false

		switch targetPhase {
		case sqlc.PhaseDEPLOY:
			if gameState.DeployableTroops == 0 {
				ctx.Log().Infow(
					"deploy must advance",
					"phase",
					gameState.Phase,
				)

				targetPhase = sqlc.PhaseATTACK
				mustAdvance = true
			}
		case sqlc.PhaseATTACK:
			targetPhase, err := s.getTargetPhaseForAttack(ctx, querier)
			if err != nil {
				return targetPhase, fmt.Errorf("failed to get target phase for attack: %w", err)
			}

			if targetPhase != sqlc.PhaseATTACK {
				mustAdvance = true
			}
		case sqlc.PhaseCONQUER:
		case sqlc.PhaseREINFORCE:
		case sqlc.PhaseCARDS:
		}
	}

	return targetPhase, nil
}

func (s *ServiceImpl) getTargetPhaseForAttack(
	ctx ctx.MoveContext,
	querier db.Querier,
) (sqlc.Phase, error) {
	targetPhase := sqlc.PhaseATTACK

	hasConquered, err := s.attackService.HasConqueredQ(ctx, querier)
	if err != nil {
		return targetPhase, fmt.Errorf("failed to check if attack has conquered: %w", err)
	}

	if hasConquered {
		ctx.Log().Infow("cannot continue attacking, must advance phase to CONQUER")

		return sqlc.PhaseCONQUER, nil
	}

	canContinueAttacking, err := s.attackService.CanContinueAttackingQ(ctx, querier)
	if err != nil {
		return targetPhase, fmt.Errorf("failed to check if attack can continue: %w", err)
	}

	if !canContinueAttacking {
		ctx.Log().Infow("cannot continue attacking, must advance phase to REINFORCE")

		return sqlc.PhaseREINFORCE, nil
	}

	return targetPhase, nil
}
