package attack

import (
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/validation"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// minTroopsToAttack is the minimum number of troops that must be sent in an attack.
	minTroopsToAttack = 1
	// minTroopsToDefend is the minimum troops a defending region must have.
	// A region with fewer troops indicates a server bug (conquered but not processed).
	minTroopsToDefend = 1
	// maxDefendingDice is the maximum number of dice the defender can roll.
	maxDefendingDice = 3
)

func (s *service) Perform(
	ctx ctx.GameContext,
	querier db.Querier,
	move Move,
	_ *snapshot.CachedGameState,
) (*MoveResult, moveservice.MoveEffect, error) {
	var zero moveservice.MoveEffect

	observe.SpanEvent(ctx, "performing_attack_move",
		attribute.String("attacking_region_id", move.AttackingRegionID),
		attribute.String("defending_region_id", move.DefendingRegionID),
		attribute.Int64("attacking_troops", move.AttackingTroops),
	)

	attackingRegion, err := s.regionService.GetRegion(ctx, querier, move.AttackingRegionID)
	if err != nil {
		return nil, zero, fmt.Errorf("unable to get attacking region: %w", err)
	}

	defendingRegion, err := s.regionService.GetRegion(ctx, querier, move.DefendingRegionID)
	if err != nil {
		return nil, zero, fmt.Errorf("unable to get defending region: %w", err)
	}

	if err := s.validate(ctx, attackingRegion, defendingRegion, move); err != nil {
		return nil, zero, fmt.Errorf("validation failed: %w", err)
	}

	casualties, err := s.perform(ctx, querier, attackingRegion, defendingRegion, move)
	if err != nil {
		return nil, zero, fmt.Errorf("unable to perform attack move: %w", err)
	}

	result := &MoveResult{
		AttackingRegionID: move.AttackingRegionID,
		DefendingRegionID: move.DefendingRegionID,
		ConqueringTroops:  move.AttackingTroops - casualties.attacking,
	}

	effect := moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{
				RegionID:  attackingRegion.ExternalReference,
				NewOwner:  attackingRegion.UserID,
				NewTroops: attackingRegion.Troops - casualties.attacking,
			},
			{
				RegionID:  defendingRegion.ExternalReference,
				NewOwner:  defendingRegion.UserID,
				NewTroops: defendingRegion.Troops - casualties.defending,
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	return result, effect, nil
}

func (s *service) perform(
	ctx ctx.GameContext,
	querier db.Querier,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) (*casualties, error) {
	attackDices := s.diceService.RollAttackingDices(int(move.AttackingTroops))
	defenseDices := s.diceService.RollDefendingDices(
		int(min(defendingRegion.Troops, maxDefendingDice)),
	)

	observe.SpanEvent(ctx, "rolled_dices",
		attribute.IntSlice("attack_dices", attackDices),
		attribute.IntSlice("defense_dices", defenseDices),
	)

	casualties := computeCasualties(ctx, attackDices, defenseDices)

	if err := s.regionService.UpdateTroopsInRegion(
		ctx,
		querier,
		attackingRegion,
		-casualties.attacking,
	); err != nil {
		return nil, fmt.Errorf("failed to decrease troops in attacking region: %w", err)
	}

	if err := s.regionService.UpdateTroopsInRegion(
		ctx,
		querier,
		defendingRegion,
		-casualties.defending,
	); err != nil {
		return nil, fmt.Errorf("failed to decrease troops in defending region: %w", err)
	}

	return casualties, nil
}

type casualties struct {
	attacking int64
	defending int64
}

func computeCasualties(ctx ctx.GameContext, attackDices, defenseDices []int) *casualties {
	casualties := &casualties{}

	slices.SortFunc(attackDices, descending)
	slices.SortFunc(defenseDices, descending)

	matches := min(len(attackDices), len(defenseDices))
	for i := range matches {
		if attackDices[i] > defenseDices[i] {
			casualties.defending++
		} else {
			casualties.attacking++
		}
	}

	observe.SpanEvent(ctx, "casualties",
		attribute.Int64("attacking", casualties.attacking),
		attribute.Int64("defending", casualties.defending),
	)

	if total := casualties.attacking + casualties.defending; total != int64(matches) {
		panic(fmt.Sprintf(
			"invariant violation: total casualties %d must equal dice matches %d",
			total, matches,
		))
	}

	if casualties.attacking < 0 || casualties.attacking > int64(len(attackDices)) {
		panic(fmt.Sprintf(
			"invariant violation: attacking casualties %d out of bounds [0, %d]",
			casualties.attacking, len(attackDices),
		))
	}

	if casualties.defending < 0 || casualties.defending > int64(len(defenseDices)) {
		panic(fmt.Sprintf(
			"invariant violation: defending casualties %d out of bounds [0, %d]",
			casualties.defending, len(defenseDices),
		))
	}

	return casualties
}

func descending(a, b int) int {
	return b - a
}

func (s *service) validate(
	ctx ctx.GameContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	if err := checkRegionOwnership(ctx, attackingRegion, defendingRegion); err != nil {
		return fmt.Errorf("region ownership check failed: %w", err)
	}

	if err := checkTroops(attackingRegion, defendingRegion, move); err != nil {
		observe.Warn(ctx, "declared troops mismatch",
			attribute.String("move_type", "attack"),
			attribute.String("source_region", attackingRegion.ExternalReference),
			attribute.String("target_region", defendingRegion.ExternalReference),
			attribute.Int64("actual_source", attackingRegion.Troops),
			attribute.Int64("actual_target", defendingRegion.Troops),
			attribute.Int64("declared_source", move.TroopsInSource),
			attribute.Int64("declared_target", move.TroopsInTarget),
		)

		return fmt.Errorf("troops check failed: %w", err)
	}

	areNeighbours, err := s.boardService.AreNeighbours(
		ctx,
		attackingRegion.ExternalReference,
		defendingRegion.ExternalReference,
	)
	if err != nil {
		return fmt.Errorf("unable to check if regions are neighbours: %w", err)
	}

	if !areNeighbours {
		return domainerrors.NewValidationError("attacking region cannot reach defending region")
	}

	return nil
}

func checkTroops(
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	if move.AttackingTroops < minTroopsToAttack {
		return domainerrors.NewValidationError("at least one troop is required to attack")
	}

	if attackingRegion.Troops <= move.AttackingTroops {
		return domainerrors.NewValidationError("attacking region does not have enough troops")
	}

	if defendingRegion.Troops < minTroopsToDefend {
		return domainerrors.NewValidationError("defending region does not have enough troops")
	}

	if err := validation.CheckDeclaredTroops(
		attackingRegion.Troops,
		defendingRegion.Troops,
		move.TroopsInSource,
		move.TroopsInTarget,
	); err != nil {
		return fmt.Errorf("declared values are invalid: %w", err)
	}

	return nil
}

func checkRegionOwnership(
	ctx ctx.GameContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
) error {
	if err := validation.CheckSourceOwnedByPlayer(ctx, attackingRegion, "attacking"); err != nil {
		return err
	}

	if err := validation.CheckTargetNotOwnedByPlayer(ctx, defendingRegion); err != nil {
		return err
	}

	return nil
}
