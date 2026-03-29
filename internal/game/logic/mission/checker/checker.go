package checker

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

// MissionChecker checks whether a specific mission type has been accomplished.
type MissionChecker interface {
	Type() sqlc.GameMissionType
	Check(
		ctx ctx.GameContext,
		querier db.Querier,
		baseMission sqlc.GameMission,
	) (bool, error)
}

// Registry maps mission types to their checkers.
type Registry struct {
	checkers map[sqlc.GameMissionType]MissionChecker
}

// NewRegistry creates a Registry from a slice of MissionCheckers.
func NewRegistry(checkers []MissionChecker) (*Registry, error) {
	checkerMap := make(map[sqlc.GameMissionType]MissionChecker, len(checkers))

	for _, c := range checkers {
		if _, exists := checkerMap[c.Type()]; exists {
			return nil, fmt.Errorf("duplicate checker registered for mission type: %s", c.Type())
		}

		checkerMap[c.Type()] = c
	}

	return &Registry{checkers: checkerMap}, nil
}

// GetChecker returns the checker for the given mission type.
func (r *Registry) GetChecker(missionType sqlc.GameMissionType) (MissionChecker, error) {
	c, ok := r.checkers[missionType]
	if !ok {
		return nil, fmt.Errorf("unknown mission type: %s", missionType)
	}

	return c, nil
}
