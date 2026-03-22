package reinforce

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	_ any,
) error {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeREINFORCE, targetPhase); err != nil {
		return fmt.Errorf("invalid phase transition: %w", err)
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

	if targetPhase == sqlc.GamePhaseTypeDEPLOY {
		if err := s.cardsService.AdvanceQ(ctx, querier, targetPhase, nil); err != nil {
			return fmt.Errorf("failed to advance cards phase: %w", err)
		}

		return nil
	}

	if _, err = s.phaseService.InsertPhaseQ(ctx, querier, sqlc.GamePhaseTypeCARDS); err != nil {
		return fmt.Errorf("failed to create cards phase: %w", err)
	}

	return nil
}
