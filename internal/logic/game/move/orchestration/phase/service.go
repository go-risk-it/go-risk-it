package phase

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/deploy"
)

type Service interface {
	AdvanceQ(ctx ctx.MoveContext, querier db.Querier) error
	SetGamePhaseQ(ctx ctx.MoveContext, querier db.Querier, phase sqlc.Phase) error
}

type ServiceImpl struct {
	attackService attack.Service
	deployService deploy.Service
	gameService   gamestate.Service
}

var _ Service = &ServiceImpl{}

func NewService(
	attackService attack.Service,
	deployService deploy.Service,
	gameService gamestate.Service,
) *ServiceImpl {
	return &ServiceImpl{
		attackService: attackService,
		deployService: deployService,
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
	gameState, err := s.gameService.GetGameStateQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	ctx.Log().Infow("walking to target phase", "from", gameState.Phase)

	targetPhase := s.walkToTargetPhase(ctx, querier, gameState)
	if targetPhase == gameState.Phase {
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
) sqlc.Phase {
	targetPhase := gameState.Phase

	mustAdvance := true
	for mustAdvance {
		mustAdvance = false

		switch targetPhase {
		case sqlc.PhaseDEPLOY:
			if s.deployService.MustAdvanceQ(ctx, querier, gameState) {
				ctx.Log().Infow(
					"deploy must advance",
					"phase",
					gameState.Phase,
				)

				targetPhase = sqlc.PhaseATTACK
				mustAdvance = true
			}
		case sqlc.PhaseATTACK:
			if s.attackService.MustAdvanceQ(ctx, querier, gameState) {
				ctx.Log().Infow(
					"attack must advance",
					"phase",
					gameState.Phase,
				)

				targetPhase = sqlc.PhaseREINFORCE
				mustAdvance = true
			}
		case sqlc.PhaseREINFORCE:
		case sqlc.PhaseCARDS:
		}
	}

	return targetPhase
}