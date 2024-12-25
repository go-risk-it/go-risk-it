package reinforce

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.PhaseType,
	_ *MoveResult,
) error {
	if targetPhase == sqlc.PhaseTypeDEPLOY {
		if err := s.cardsService.AdvanceQ(ctx, querier, targetPhase, nil); err != nil {
			return fmt.Errorf("failed to advance cards phase: %w", err)
		}

		return nil
	}

	if targetPhase != sqlc.PhaseTypeCARDS {
		return fmt.Errorf("cannot advance reinforce phase to %s", targetPhase)
	}

	_, err := s.phaseService.InsertPhaseQ(ctx, querier, sqlc.PhaseTypeCARDS)
	if err != nil {
		return fmt.Errorf("failed to create cards phase: %w", err)
	}

	game, err := s.gameService.GetGameStateQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("unable to get game state: %w", err)
	}

	hasConqueredInTurn, err := querier.HasConqueredInTurn(ctx, sqlc.HasConqueredInTurnParams{
		ID:   ctx.GameID(),
		Turn: game.Turn - 1,
	})
	if err != nil {
		return fmt.Errorf("failed to check if player has conquered in turn: %w", err)
	}

	if hasConqueredInTurn {
		if err := s.cardsService.Draw(ctx, querier); err != nil {
			return fmt.Errorf("failed to draw cards: %w", err)
		}
	}

	return nil
}
