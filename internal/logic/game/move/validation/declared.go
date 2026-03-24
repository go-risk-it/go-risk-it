package validation

import (
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
)

// CheckDeclaredTroops validates that the declared troop counts match actual region troops.
func CheckDeclaredTroops(
	ctx ctx.GameContext,
	sourceTroops int64,
	targetTroops int64,
	declaredSourceTroops int64,
	declaredTargetTroops int64,
) error {
	slog.InfoContext(ctx, "checking declared values")

	if sourceTroops != declaredSourceTroops {
		return domainerrors.NewValidationError(
			"source region doesn't have the declared number of troops",
		)
	}

	if targetTroops != declaredTargetTroops {
		return domainerrors.NewValidationError(
			"target region doesn't have the declared number of troops",
		)
	}

	slog.InfoContext(ctx, "declared values check passed")

	return nil
}
