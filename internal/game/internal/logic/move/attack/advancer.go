package attack

import (
	"errors"
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
	performResult *MoveResult,
	_ moveservice.AdvanceContext,
) (moveservice.AdvanceEffect, error) {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeATTACK, targetPhase); err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("invalid phase transition: %w", err)
	}

	dbPhase, err := s.phaseService.InsertPhase(ctx, querier, targetPhase)
	if err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("failed to create phase: %w", err)
	}

	if targetPhase == sqlc.GamePhaseTypeCONQUER {
		effect, err := s.advanceToConquerPhase(ctx, querier, *dbPhase, performResult)
		if err != nil {
			return moveservice.AdvanceEffect{}, err
		}

		return effect, nil
	}

	return moveservice.AdvanceEffect{
		NewPhase:  snapshot.EmptyPhaseState{},
		TurnEnded: false,
	}, nil
}

func (s *service) advanceToConquerPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	dbPhase sqlc.GamePhase,
	performResult *MoveResult,
) (moveservice.AdvanceEffect, error) {
	if performResult == nil {
		return moveservice.AdvanceEffect{}, errors.New(
			"no attack result available for conquer phase creation",
		)
	}

	if _, err := querier.InsertConquerPhase(ctx, sqlc.InsertConquerPhaseParams{
		PhaseID:             dbPhase.ID,
		ID:                  ctx.GameID(),
		ExternalReference:   performResult.AttackingRegionID,
		ExternalReference_2: performResult.DefendingRegionID,
		MinimumTroops:       performResult.ConqueringTroops,
	}); err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("failed to create conquer phase: %w", err)
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
