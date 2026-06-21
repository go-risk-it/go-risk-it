package checker

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
)

// CheckContext holds all data needed by mission checkers to evaluate
// win conditions without any DB access.
type CheckContext struct {
	Regions       []snapshot.RegionState
	Continents    board.Continents
	CurrentUserID string
}

// MissionChecker checks whether a specific mission type has been accomplished.
// Implementations are pure evaluators — no DB access, no side effects.
type MissionChecker interface {
	Type() snapshot.MissionType
	Check(
		checkCtx CheckContext,
		mission snapshot.PlayerMission,
	) (bool, error)
}

// Registry maps mission types to their checkers.
type Registry struct {
	checkers map[snapshot.MissionType]MissionChecker
}

// NewRegistry creates a Registry from a slice of MissionCheckers.
func NewRegistry(checkers []MissionChecker) (*Registry, error) {
	checkerMap := make(map[snapshot.MissionType]MissionChecker, len(checkers))

	for _, c := range checkers {
		if _, exists := checkerMap[c.Type()]; exists {
			return nil, fmt.Errorf("duplicate checker registered for mission type: %s", c.Type())
		}

		checkerMap[c.Type()] = c
	}

	return &Registry{checkers: checkerMap}, nil
}

// GetChecker returns the checker for the given mission type.
func (r *Registry) GetChecker(missionType snapshot.MissionType) (MissionChecker, error) {
	c, ok := r.checkers[missionType]
	if !ok {
		return nil, fmt.Errorf("unknown mission type: %s", missionType)
	}

	return c, nil
}
