package validation

import (
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

// CheckSourceOwnedByPlayer validates that the source region is owned by the current player.
// The regionLabel parameter controls the error message (e.g., "attacking" or "source").
func CheckSourceOwnedByPlayer(
	ctx ctx.GameContext,
	region *sqlc.GetRegionsByGameRow,
	regionLabel string,
) error {
	if region.UserID != ctx.UserID() {
		return domainerrors.NewValidationError(
			regionLabel + " region is not owned by player",
		)
	}

	return nil
}

// CheckTargetNotOwnedByPlayer validates that the target region is NOT owned by the current player.
// Used for attack moves where you cannot attack your own region.
func CheckTargetNotOwnedByPlayer(
	ctx ctx.GameContext,
	region *sqlc.GetRegionsByGameRow,
) error {
	if region.UserID == ctx.UserID() {
		return domainerrors.NewValidationError("cannot attack your own region")
	}

	return nil
}

// CheckTargetOwnedByPlayer validates that the target region IS owned by the current player.
// Used for reinforce moves where both regions must belong to the player.
func CheckTargetOwnedByPlayer(
	ctx ctx.GameContext,
	region *sqlc.GetRegionsByGameRow,
) error {
	if region.UserID != ctx.UserID() {
		return domainerrors.NewValidationError("target region is not owned by player")
	}

	return nil
}
