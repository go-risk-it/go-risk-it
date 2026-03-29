package reinforce

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/phase"
)

func (s *service) Advance(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	_ struct{},
) error {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeREINFORCE, targetPhase); err != nil {
		return fmt.Errorf("invalid phase transition: %w", err)
	}

	game, err := s.gameService.GetGameStateWithQuerier(ctx, querier)
	if err != nil {
		return fmt.Errorf("unable to get game state: %w", err)
	}

	slog.DebugContext(ctx, "checking if player has conquered", "turn", game.Turn)

	hasConqueredInTurn, err := querier.HasConqueredInTurn(ctx, sqlc.HasConqueredInTurnParams{
		ID:   ctx.GameID(),
		Turn: game.Turn,
	})
	if err != nil {
		return fmt.Errorf("failed to check if player has conquered in turn: %w", err)
	}

	if hasConqueredInTurn {
		slog.InfoContext(ctx, "player has conquered in turn")

		if err := s.cardsService.Draw(ctx, querier); err != nil {
			return fmt.Errorf("failed to draw cards: %w", err)
		}
	}

	if targetPhase == sqlc.GamePhaseTypeDEPLOY {
		if err := s.cardsService.Advance(ctx, querier, targetPhase, nil); err != nil {
			return fmt.Errorf("failed to advance cards phase: %w", err)
		}

		return nil
	}

	if _, err = s.phaseService.InsertPhase(ctx, querier, sqlc.GamePhaseTypeCARDS); err != nil {
		return fmt.Errorf("failed to create cards phase: %w", err)
	}

	return nil
}
