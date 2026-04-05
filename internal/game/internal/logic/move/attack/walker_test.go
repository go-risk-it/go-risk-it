package attack_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/attack"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/move/attack/dice"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/region"
	"github.com/stretchr/testify/require"
)

func TestWalk(t *testing.T) {
	t.Parallel()

	const currentUser = "player1"

	tests := []struct {
		name          string
		voluntary     bool
		prevRegions   []snapshot.RegionState
		regionUpdates []moveservice.RegionUpdate
		expectedPhase sqlc.GamePhaseType
	}{
		{
			name:      "conquered region (0 troops, different owner) returns CONQUER",
			voluntary: false,
			prevRegions: []snapshot.RegionState{
				{ID: "r1", OwnerID: currentUser, Troops: 5},
				{ID: "r2", OwnerID: "enemy", Troops: 3},
			},
			regionUpdates: []moveservice.RegionUpdate{
				{RegionID: "r2", NewOwner: "enemy", NewTroops: 0},
			},
			expectedPhase: sqlc.GamePhaseTypeCONQUER,
		},
		{
			name:      "no conquered region, can continue attacking returns ATTACK",
			voluntary: false,
			prevRegions: []snapshot.RegionState{
				{ID: "r1", OwnerID: currentUser, Troops: 5},
				{ID: "r2", OwnerID: "enemy", Troops: 1},
			},
			regionUpdates: nil,
			expectedPhase: sqlc.GamePhaseTypeATTACK,
		},
		{
			name:      "no conquered region, cannot continue attacking returns REINFORCE",
			voluntary: false,
			prevRegions: []snapshot.RegionState{
				{ID: "r1", OwnerID: currentUser, Troops: 1},
				{ID: "r2", OwnerID: "enemy", Troops: 3},
			},
			regionUpdates: nil,
			expectedPhase: sqlc.GamePhaseTypeREINFORCE,
		},
		{
			name:      "voluntary advancement returns REINFORCE even if can continue",
			voluntary: true,
			prevRegions: []snapshot.RegionState{
				{ID: "r1", OwnerID: currentUser, Troops: 5},
				{ID: "r2", OwnerID: "enemy", Troops: 3},
			},
			regionUpdates: nil,
			expectedPhase: sqlc.GamePhaseTypeREINFORCE,
		},
		{
			name:      "region update applies over prev snapshot",
			voluntary: false,
			prevRegions: []snapshot.RegionState{
				{ID: "r1", OwnerID: currentUser, Troops: 3},
				{ID: "r2", OwnerID: "enemy", Troops: 5},
			},
			regionUpdates: []moveservice.RegionUpdate{
				// Attack reduced source and zeroed target.
				{RegionID: "r1", NewOwner: currentUser, NewTroops: 1},
				{RegionID: "r2", NewOwner: "enemy", NewTroops: 0},
			},
			expectedPhase: sqlc.GamePhaseTypeCONQUER,
		},
		{
			name:      "all player regions have exactly 1 troop returns REINFORCE",
			voluntary: false,
			prevRegions: []snapshot.RegionState{
				{ID: "r1", OwnerID: currentUser, Troops: 1},
				{ID: "r2", OwnerID: currentUser, Troops: 1},
				{ID: "r3", OwnerID: "enemy", Troops: 5},
			},
			regionUpdates: nil,
			expectedPhase: sqlc.GamePhaseTypeREINFORCE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Construct a real service (deps unused by Walk).
			svc, _ := attack.NewService(
				board.NewService(t),
				dice.NewService(t),
				phase.NewService(t),
				region.NewService(t),
			)

			wctx := moveservice.WalkContext{
				Voluntary: tt.voluntary,
				PrevSnapshot: &snapshot.GameSnapshot{
					Regions: tt.prevRegions,
				},
				Effect: moveservice.MoveEffect{
					RegionUpdates: tt.regionUpdates,
				},
				CurrentUserID: currentUser,
			}

			got, err := svc.Walk(wctx)

			require.NoError(t, err)
			require.Equal(t, tt.expectedPhase, got)
		})
	}
}
