package conquer_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/conquer"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/card"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/mission"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/region"
	"github.com/stretchr/testify/require"
)

func TestWalk(t *testing.T) {
	t.Parallel()

	const currentUser = "player1"

	tests := []struct {
		name          string
		prevRegions   []snapshot.RegionState
		regionUpdates []moveservice.RegionUpdate
		expectedPhase sqlc.GamePhaseType
	}{
		{
			name: "can continue attacking returns ATTACK",
			prevRegions: []snapshot.RegionState{
				{ID: "r1", OwnerID: currentUser, Troops: 5},
				{ID: "r2", OwnerID: currentUser, Troops: 2},
				{ID: "r3", OwnerID: "enemy", Troops: 3},
			},
			regionUpdates: nil,
			expectedPhase: sqlc.GamePhaseTypeATTACK,
		},
		{
			name: "cannot continue attacking returns REINFORCE",
			prevRegions: []snapshot.RegionState{
				{ID: "r1", OwnerID: currentUser, Troops: 1},
				{ID: "r2", OwnerID: currentUser, Troops: 1},
				{ID: "r3", OwnerID: "enemy", Troops: 3},
			},
			regionUpdates: nil,
			expectedPhase: sqlc.GamePhaseTypeREINFORCE,
		},
		{
			name: "region updates make attacking possible returns ATTACK",
			prevRegions: []snapshot.RegionState{
				{ID: "r1", OwnerID: currentUser, Troops: 1},
				{ID: "r2", OwnerID: "enemy", Troops: 3},
			},
			regionUpdates: []moveservice.RegionUpdate{
				// Conquer moved troops into conquered region.
				{RegionID: "r1", NewOwner: currentUser, NewTroops: 1},
				{RegionID: "r2", NewOwner: currentUser, NewTroops: 4},
			},
			expectedPhase: sqlc.GamePhaseTypeATTACK,
		},
		{
			name: "region updates leave all troops at 1 returns REINFORCE",
			prevRegions: []snapshot.RegionState{
				{ID: "r1", OwnerID: currentUser, Troops: 3},
				{ID: "r2", OwnerID: "enemy", Troops: 5},
			},
			regionUpdates: []moveservice.RegionUpdate{
				{RegionID: "r1", NewOwner: currentUser, NewTroops: 1},
				{RegionID: "r2", NewOwner: currentUser, NewTroops: 1},
			},
			expectedPhase: sqlc.GamePhaseTypeREINFORCE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Construct a real service (deps unused by Walk).
			svc, _ := conquer.NewService(
				db.NewQuerier(t),
				attack.NewService(t),
				card.NewService(t),
				mission.NewService(t),
				phase.NewService(t),
				region.NewService(t),
			)

			wctx := moveservice.WalkContext{
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
