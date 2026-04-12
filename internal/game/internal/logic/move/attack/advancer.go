package attack

import (
	"errors"
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
	performResult *MoveResult,
	advCtx moveservice.AdvanceContext,
) (moveservice.AdvanceEffect, error) {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeATTACK, targetPhase); err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("invalid phase transition: %w", err)
	}

	if targetPhase == sqlc.GamePhaseTypeCONQUER {
		if performResult == nil {
			return moveservice.AdvanceEffect{}, errors.New(
				"no attack result available for conquer phase creation",
			)
		}

		return moveservice.AdvanceEffect{
			NewPhase: snapshot.ConquerPhaseState{
				AttackingRegionID: performResult.AttackingRegionID,
				DefendingRegionID: performResult.DefendingRegionID,
				MinTroopsToMove:   performResult.ConqueringTroops,
			},
			TurnEnded: false,
		}, nil
	}

	return moveservice.AdvanceEffect{
		NewPhase:  snapshot.EmptyPhaseState{},
		TurnEnded: false,
	}, nil
}
