package conquer

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
	_ moveservice.AdvanceContext,
) (moveservice.AdvanceEffect, error) {
	if err := phase.ValidateTransition(sqlc.GamePhaseTypeCONQUER, targetPhase); err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("invalid phase transition: %w", err)
	}

	_, err := s.phaseService.InsertPhase(ctx, querier, targetPhase)
	if err != nil {
		return moveservice.AdvanceEffect{}, fmt.Errorf("failed to create phase: %w", err)
	}

	return moveservice.AdvanceEffect{
		NewPhase:  snapshot.EmptyPhaseState{},
		TurnEnded: false,
	}, nil
}
