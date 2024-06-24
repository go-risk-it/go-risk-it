package phase

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
	"go.uber.org/zap"
)

type Service interface {
	AdvanceQ(ctx riskcontext.MoveContext, querier db.Querier) error
	SetGamePhaseQ(ctx riskcontext.MoveContext, querier db.Querier, phase sqlc.Phase) error
}

type ServiceImpl struct {
	log           *zap.SugaredLogger
	attackService attack.Service
	deployService deploy.Service
	gameService   game.Service
}

var _ Service = &ServiceImpl{}

func NewService(
	log *zap.SugaredLogger,
	attackService attack.Service,
	deployService deploy.Service,
	gameService game.Service,
) *ServiceImpl {
	return &ServiceImpl{
		log:           log,
		attackService: attackService,
		deployService: deployService,
		gameService:   gameService,
	}
}

func (s *ServiceImpl) SetGamePhaseQ(
	ctx riskcontext.MoveContext,
	querier db.Querier,
	phase sqlc.Phase,
) error {
	s.log.Infow("setting phase", "gameID", ctx.GameID(), "phase", phase)

	err := querier.SetGamePhase(ctx, sqlc.SetGamePhaseParams{
		Phase: phase,
		ID:    ctx.GameID(),
	})
	if err != nil {
		return fmt.Errorf("failed to set phase: %w", err)
	}

	s.log.Infow("phase set", "gameID", ctx.GameID(), "phase", phase)

	return nil
}

func (s *ServiceImpl) AdvanceQ(ctx riskcontext.MoveContext, querier db.Querier) error {
	gameState, err := s.gameService.GetGameStateQ(ctx, querier, ctx.GameID())
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	s.log.Infow("Walking to target phase", "gameID", ctx.GameID(), "from", gameState.Phase)

	targetPhase := s.walkToTargetPhase(ctx, querier, gameState)
	if targetPhase == gameState.Phase {
		return nil
	}

	s.log.Infow(
		"Advancing phase",
		"gameID",
		ctx.GameID(),
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
	ctx riskcontext.MoveContext,
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
				s.log.Infow(
					"deploy must advance",
					"gameID",
					gameState.ID,
					"phase",
					gameState.Phase,
				)

				targetPhase = sqlc.PhaseATTACK
				mustAdvance = true
			}
		case sqlc.PhaseATTACK:
			if s.attackService.MustAdvanceQ(ctx, querier, gameState) {
				s.log.Infow(
					"attack must advance",
					"gameID",
					gameState.ID,
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
