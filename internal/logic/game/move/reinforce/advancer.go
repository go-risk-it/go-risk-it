package reinforce

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.PhaseType,
	_ *MoveResult,
) error {
	if targetPhase != sqlc.PhaseTypeDEPLOY && targetPhase != sqlc.PhaseTypeCARDS {
		return fmt.Errorf("cannot advance reinforce phase to %s", targetPhase)
	}

	game, err := s.gameService.GetGameStateQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("unable to get game state: %w", err)
	}

	ctx.Log().Debugf("checking if player has conquered in turn %d", game.Turn)

	hasConqueredInTurn, err := querier.HasConqueredInTurn(ctx, sqlc.HasConqueredInTurnParams{
		ID:   ctx.GameID(),
		Turn: game.Turn,
	})
	if err != nil {
		return fmt.Errorf("failed to check if player has conquered in turn: %w", err)
	}

	if hasConqueredInTurn {
		ctx.Log().Infow("player has conquered in turn")

		if err := s.cardsService.Draw(ctx, querier); err != nil {
			return fmt.Errorf("failed to draw cards: %w", err)
		}
	}

	if targetPhase == sqlc.PhaseTypeDEPLOY {
		if err := s.cardsService.AdvanceQ(ctx, querier, targetPhase, nil); err != nil {
			return fmt.Errorf("failed to advance cards phase: %w", err)
		}

		return nil
	}

	if _, err = s.phaseService.InsertPhaseQ(ctx, querier, sqlc.PhaseTypeCARDS); err != nil {
		return fmt.Errorf("failed to create cards phase: %w", err)
	}

	return nil
}
