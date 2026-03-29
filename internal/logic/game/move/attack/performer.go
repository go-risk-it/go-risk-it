package attack

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/validation"
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
) (*MoveResult, error) {
	slog.DebugContext(ctx, "performing attack move", "move", move)

	attackingRegion, err := s.regionService.GetRegion(ctx, querier, move.AttackingRegionID)
	if err != nil {
		return nil, fmt.Errorf("unable to get attacking region: %w", err)
	}

	defendingRegion, err := s.regionService.GetRegion(ctx, querier, move.DefendingRegionID)
	if err != nil {
		return nil, fmt.Errorf("unable to get defending region: %w", err)
	}

	if err := s.validate(ctx, attackingRegion, defendingRegion, move); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	casualties, err := s.perform(ctx, querier, attackingRegion, defendingRegion, move)
	if err != nil {
		return nil, fmt.Errorf("unable to perform attack move: %w", err)
	}

	slog.DebugContext(ctx, "attack executed successfully")

	result := &MoveResult{
		AttackingRegionID: move.AttackingRegionID,
		DefendingRegionID: move.DefendingRegionID,
		ConqueringTroops:  move.AttackingTroops - casualties.attacking,
	}

	return result, nil
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

	slog.DebugContext(ctx, "rolled dices", "attack", attackDices, "defense", defenseDices)

	casualties := computeCasualties(ctx, attackDices, defenseDices)

	slog.DebugContext(ctx, "updating region troops")

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

	slog.DebugContext(ctx,
		"casualties",
		"attacking",
		casualties.attacking,
		"defending",
		casualties.defending)

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
	slog.DebugContext(ctx, "validating attack move", "move", move)

	if err := checkRegionOwnership(ctx, attackingRegion, defendingRegion); err != nil {
		return fmt.Errorf("region ownership check failed: %w", err)
	}

	if err := checkTroops(ctx, attackingRegion, defendingRegion, move); err != nil {
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

	slog.DebugContext(ctx, "attack move validation passed", "move", move)

	return nil
}

func checkTroops(
	ctx ctx.GameContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
	move Move,
) error {
	slog.DebugContext(ctx, "checking troops")

	if move.AttackingTroops < minTroopsToAttack {
		return domainerrors.NewValidationError("at least one troop is required to attack")
	}

	if attackingRegion.Troops <= move.AttackingTroops {
		return domainerrors.NewValidationError("attacking region does not have enough troops")
	}

	if defendingRegion.Troops < minTroopsToDefend {
		slog.ErrorContext(ctx,
			"attempting to attack a region with no troops — possible server bug",
			"defendingRegion", defendingRegion.ExternalReference,
			"defendingTroops", defendingRegion.Troops,
			"defendingOwner", defendingRegion.UserID,
			"attackingRegion", attackingRegion.ExternalReference,
			"attackingTroops", attackingRegion.Troops,
			"attackingOwner", attackingRegion.UserID,
			"moveAttackingTroops", move.AttackingTroops,
			"gameID", ctx.GameID(),
			"userID", ctx.UserID(),
		)

		return domainerrors.NewValidationError("defending region does not have enough troops")
	}

	if err := validation.CheckDeclaredTroops(
		ctx,
		attackingRegion.Troops,
		defendingRegion.Troops,
		move.TroopsInSource,
		move.TroopsInTarget,
	); err != nil {
		return fmt.Errorf("declared values are invalid: %w", err)
	}

	slog.DebugContext(ctx, "troops check passed")

	return nil
}

func checkRegionOwnership(
	ctx ctx.GameContext,
	attackingRegion *sqlc.GetRegionsByGameRow,
	defendingRegion *sqlc.GetRegionsByGameRow,
) error {
	slog.DebugContext(ctx, "checking region ownership")

	if err := validation.CheckSourceOwnedByPlayer(ctx, attackingRegion, "attacking"); err != nil {
		return err
	}

	if err := validation.CheckTargetNotOwnedByPlayer(ctx, defendingRegion); err != nil {
		return err
	}

	slog.DebugContext(ctx, "region ownership check passed")

	return nil
}
