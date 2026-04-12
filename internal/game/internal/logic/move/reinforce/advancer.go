package reinforce

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/phase"
)

func (s *service) Advance(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.GamePhaseType,
	_ struct{},
	advCtx moveservice.AdvanceContext,
) (moveservice.AdvanceEffect, error) {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeREINFORCE, targetPhase); err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("invalid phase transition: %w", err)
	}

	// Draw a card if the player conquered a region this turn.
	var cardDeltas []moveservice.CardDelta

	if advCtx.ConqueredInTurn {
		drawnCard, err := s.cardsService.Draw(ctx, querier)
		if err != nil {
			return moveservice.AdvanceEffect{}, fmt.Errorf("failed to draw cards: %w", err)
		}

		cardDeltas = []moveservice.CardDelta{
			{
				PlayerUserID: ctx.UserID(),
				Gained:       []snapshot.CardState{drawnCard},
			},
		}
	}

	// When advancing to DEPLOY, delegate to cards.Advance to create the deploy
	// phase with the correct deployable troops calculation.
	if targetPhase == sqlc.GamePhaseTypeDEPLOY {
		cardsEffect, err := s.cardsService.Advance(ctx, querier, targetPhase, nil, advCtx)
		if err != nil {
			return moveservice.AdvanceEffect{}, fmt.Errorf(
				"failed to advance cards phase: %w",
				err,
			)
		}

		// The turn ends when advancing to DEPLOY (turn boundary).
		// Merge card deltas from the draw with whatever cards.Advance produced.
		cardDeltas = append(cardDeltas, cardsEffect.CardDeltas...)

		return moveservice.AdvanceEffect{
			NewPhase:   cardsEffect.NewPhase,
			TurnEnded:  true,
			CardDeltas: cardDeltas,
		}, nil
	}

	// Advancing to CARDS phase — the next player must play card combinations.
	if _, err := s.phaseService.InsertPhase(ctx, querier, sqlc.GamePhaseTypeCARDS); err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("failed to create cards phase: %w", err)
	}

	return moveservice.AdvanceEffect{
		NewPhase:   snapshot.EmptyPhaseState{},
		TurnEnded:  true,
		CardDeltas: cardDeltas,
	}, nil
}
