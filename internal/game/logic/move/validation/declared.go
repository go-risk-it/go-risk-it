package validation

import (
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

// CheckDeclaredTroops validates that the declared troop counts match actual region troops.
func CheckDeclaredTroops(
	sourceTroops int64,
	targetTroops int64,
	declaredSourceTroops int64,
	declaredTargetTroops int64,
) error {
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

	return nil
}
