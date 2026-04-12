package phase

import (
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
)

// ValidTransitions defines the complete phase state machine.
// Each key maps to the set of phases it is allowed to transition to.
var ValidTransitions = map[sqlc.GamePhaseType][]sqlc.GamePhaseType{
	sqlc.GamePhaseTypeCARDS:     {sqlc.GamePhaseTypeDEPLOY},
	sqlc.GamePhaseTypeDEPLOY:    {sqlc.GamePhaseTypeATTACK},
	sqlc.GamePhaseTypeATTACK:    {sqlc.GamePhaseTypeCONQUER, sqlc.GamePhaseTypeREINFORCE},
	sqlc.GamePhaseTypeCONQUER:   {sqlc.GamePhaseTypeATTACK, sqlc.GamePhaseTypeREINFORCE},
	sqlc.GamePhaseTypeREINFORCE: {sqlc.GamePhaseTypeCARDS, sqlc.GamePhaseTypeDEPLOY},
}

// ValidateTransition checks whether transitioning from → to is allowed by
// the state machine. Returns a descriptive error if the transition is invalid.
func ValidateTransition(from, target sqlc.GamePhaseType) error {
	allowed, ok := ValidTransitions[from]
	if !ok {
		return fmt.Errorf("unknown phase %q", from)
	}

	if slices.Contains(allowed, target) {
		return nil
	}

	return fmt.Errorf("cannot advance from %s to %s", from, target)
}
