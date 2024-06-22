package orchestration

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/validation"
	"github.com/go-risk-it/go-risk-it/internal/signals"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Service interface {
	PerformMove(
		ctx context.Context,
		gameID int64,
		userID string,
		validatePhase func(game *sqlc.Game) bool,
		perform func(ctx context.Context, querier db.Querier, game *sqlc.Game) error,
	) error
}

type ServiceImpl struct {
	log                      *zap.SugaredLogger
	querier                  db.Querier
	attackService            attack.Service
	deployService            deploy.Service
	gameService              game.Service
	validationService        validation.Service
	boardStateChangedSignal  signals.BoardStateChangedSignal
	playerStateChangedSignal signals.PlayerStateChangedSignal
	gameStateChangedSignal   signals.GameStateChangedSignal
}

func NewService(
	log *zap.SugaredLogger,
	querier db.Querier,
	attackService attack.Service,
	deployService deploy.Service,
	gameService game.Service,
	validationService validation.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
	playerStateChangedSignal signals.PlayerStateChangedSignal,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		log:                      log,
		querier:                  querier,
		attackService:            attackService,
		deployService:            deployService,
		gameService:              gameService,
		validationService:        validationService,
		boardStateChangedSignal:  boardStateChangedSignal,
		playerStateChangedSignal: playerStateChangedSignal,
		gameStateChangedSignal:   gameStateChangedSignal,
	}
}

func (s *ServiceImpl) PerformMove(
	ctx context.Context,
	gameID int64,
	userID string,
	validatePhase func(game *sqlc.Game) bool,
	perform func(ctx context.Context, querier db.Querier, game *sqlc.Game) error,
) error {
	_, err := s.querier.ExecuteInTransactionWithIsolation(
		ctx,
		pgx.RepeatableRead,
		func(querier db.Querier) (interface{}, error) {
			gameState, err := s.gameService.GetGameStateQ(ctx, querier, gameID)
			if err != nil {
				return nil, fmt.Errorf("unable to get game state: %w", err)
			}

			if !validatePhase(gameState) {
				return nil, fmt.Errorf("game is not in the correct phase to perform move")
			}

			if err := s.validationService.Validate(
				ctx,
				querier,
				gameState,
				userID); err != nil {
				return nil, fmt.Errorf("invalid move: %w", err)
			}

			if err := perform(ctx, querier, gameState); err != nil {
				return nil, fmt.Errorf("unable to perform move: %w", err)
			}

			if err := s.advancePhaseQ(
				ctx,
				querier,
				gameID); err != nil {
				return nil, fmt.Errorf("unable to advance phase: %w", err)
			}

			return struct{}{}, nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	s.publishMoveResult(ctx, gameID)

	return nil
}

func (s *ServiceImpl) advancePhaseQ(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
) error {
	gameState, err := s.gameService.GetGameStateQ(ctx, querier, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	s.log.Infow("Walking to target phase", "gameID", gameID, "from", gameState.Phase)

	targetPhase := s.walkToTargetPhase(ctx, querier, gameState)
	if targetPhase == gameState.Phase {
		return nil
	}

	s.log.Infow("Advancing phase", "gameID", gameID, "from", gameState.Phase, "to", targetPhase)

	err = s.gameService.SetGamePhaseQ(ctx, querier, gameID, targetPhase)
	if err != nil {
		return fmt.Errorf("failed to set game phase: %w", err)
	}

	return nil
}

func (s *ServiceImpl) publishMoveResult(ctx context.Context, gameID int64) {
	go s.boardStateChangedSignal.Emit(ctx, signals.BoardStateChangedData{
		GameID: gameID,
	})
	go s.playerStateChangedSignal.Emit(ctx, signals.PlayerStateChangedData{
		GameID: gameID,
	})
	go s.gameStateChangedSignal.Emit(ctx, signals.GameStateChangedData{
		GameID: gameID,
	})
}

func (s *ServiceImpl) walkToTargetPhase(
	ctx context.Context,
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
