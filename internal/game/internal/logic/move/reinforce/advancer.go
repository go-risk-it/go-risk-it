package reinforce

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/phase"
)

func (s *service) Advance(
	ctx ctx.GameContext,
	targetPhase sqlc.GamePhaseType,
	_ struct{},
	advCtx moveservice.AdvanceContext,
) (moveservice.AdvanceEffect, error) {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeREINFORCE, targetPhase); err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("invalid phase transition: %w", err)
	}

	// Draw a card if the player conquered a region this turn.
	var cardDeltas []moveservice.CardDelta

	var deckDelta moveservice.DeckDelta

	if advCtx.ConqueredInTurn {
		drawnCard, err := s.cardsService.Draw(advCtx.AvailableDeck)
		if err != nil {
			return moveservice.AdvanceEffect{}, fmt.Errorf("failed to draw card: %w", err)
		}

		cardDeltas = []moveservice.CardDelta{
			{
				PlayerUserID: ctx.UserID(),
				Gained:       []snapshot.CardState{drawnCard},
			},
		}

		deckDelta = moveservice.DeckDelta{
			Drawn: []int64{drawnCard.ID},
		}
	}

	// When advancing to DEPLOY, delegate to cards.Advance to create the deploy
	// phase with the correct deployable troops calculation.
	// The turn ends here (REINFORCE→DEPLOY), so advance the turn to the next
	// alive player before delegating — cards.Advance uses advCtx.Turn to
	// determine the deploy-for player.
	if targetPhase == sqlc.GamePhaseTypeDEPLOY {
		nextAdvCtx := advCtx
		nextAdvCtx.Turn = nextAliveTurn(advCtx.Turn, advCtx.Players)

		cardsEffect, err := s.cardsService.Advance(ctx, targetPhase, nil, nextAdvCtx)
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
			DeckDelta:  deckDelta,
		}, nil
	}

	// Advancing to CARDS phase — the next player must play card combinations.
	return moveservice.AdvanceEffect{
		NewPhase:   snapshot.EmptyPhaseState{},
		TurnEnded:  true,
		CardDeltas: cardDeltas,
		DeckDelta:  deckDelta,
	}, nil
}

// nextAliveTurn finds the next turn index (wrapping around) that corresponds
// to an alive player. Mirrors orchestration.nextAliveTurn but is local to
// avoid a cross-package dependency.
func nextAliveTurn(
	currentTurn int64,
	players []snapshot.PlayerState,
) int64 {
	count := int64(len(players))
	if count == 0 {
		return currentTurn + 1
	}

	statusByIndex := make(map[int64]snapshot.PlayerStatus, len(players))
	for _, p := range players {
		statusByIndex[p.Index] = p.Status
	}

	turn := currentTurn + 1

	for range count {
		if statusByIndex[turn%count] == snapshot.PlayerAlive {
			return turn
		}

		turn++
	}

	return currentTurn + 1
}
