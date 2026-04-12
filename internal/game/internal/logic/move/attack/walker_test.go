package attack_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/attack"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/testmocks/logic/move/attack/dice"
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
				{InternalID: 1, ID: "r1", OwnerID: currentUser, Troops: 5},
				{InternalID: 2, ID: "r2", OwnerID: "enemy", Troops: 3},
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
				{InternalID: 1, ID: "r1", OwnerID: currentUser, Troops: 5},
				{InternalID: 2, ID: "r2", OwnerID: "enemy", Troops: 1},
			},
			regionUpdates: nil,
			expectedPhase: sqlc.GamePhaseTypeATTACK,
		},
		{
			name:      "no conquered region, cannot continue attacking returns REINFORCE",
			voluntary: false,
			prevRegions: []snapshot.RegionState{
				{InternalID: 1, ID: "r1", OwnerID: currentUser, Troops: 1},
				{InternalID: 2, ID: "r2", OwnerID: "enemy", Troops: 3},
			},
			regionUpdates: nil,
			expectedPhase: sqlc.GamePhaseTypeREINFORCE,
		},
		{
			name:      "voluntary advancement returns REINFORCE even if can continue",
			voluntary: true,
			prevRegions: []snapshot.RegionState{
				{InternalID: 1, ID: "r1", OwnerID: currentUser, Troops: 5},
				{InternalID: 2, ID: "r2", OwnerID: "enemy", Troops: 3},
			},
			regionUpdates: nil,
			expectedPhase: sqlc.GamePhaseTypeREINFORCE,
		},
		{
			name:      "region update applies over prev snapshot",
			voluntary: false,
			prevRegions: []snapshot.RegionState{
				{InternalID: 1, ID: "r1", OwnerID: currentUser, Troops: 3},
				{InternalID: 2, ID: "r2", OwnerID: "enemy", Troops: 5},
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
				{InternalID: 1, ID: "r1", OwnerID: currentUser, Troops: 1},
				{InternalID: 2, ID: "r2", OwnerID: currentUser, Troops: 1},
				{InternalID: 3, ID: "r3", OwnerID: "enemy", Troops: 5},
			},
			regionUpdates: nil,
			expectedPhase: sqlc.GamePhaseTypeREINFORCE,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Construct a real service (deps unused by Walk).
			svc, _ := attack.NewService(
				board.NewService(t),
				dice.NewService(t),
			)

			wctx := moveservice.WalkContext{
				Voluntary: testCase.voluntary,
				PrevSnapshot: &snapshot.GameSnapshot{
					Regions: testCase.prevRegions,
				},
				Effect: moveservice.MoveEffect{
					RegionUpdates: testCase.regionUpdates,
				},
				CurrentUserID: currentUser,
			}

			got, err := svc.Walk(wctx)

			require.NoError(t, err)
			require.Equal(t, testCase.expectedPhase, got)
		})
	}
}

func TestMergeRegions_PreservesInternalID(t *testing.T) {
	t.Parallel()

	base := []snapshot.RegionState{
		{InternalID: 10, ID: "r1", OwnerID: "player1", Troops: 5},
		{InternalID: 20, ID: "r2", OwnerID: "enemy", Troops: 3},
		{InternalID: 30, ID: "r3", OwnerID: "player1", Troops: 2},
	}

	updates := []moveservice.RegionUpdate{
		{RegionID: "r1", NewOwner: "player1", NewTroops: 2},
		{RegionID: "r2", NewOwner: "player1", NewTroops: 3},
	}

	result := attack.MergeRegions(base, updates)

	require.Len(t, result, 3)

	// Updated regions must preserve InternalID from base.
	require.Equal(t, int64(10), result[0].InternalID, "r1 InternalID preserved")
	require.Equal(t, int64(2), result[0].Troops)
	require.Equal(t, "player1", result[0].OwnerID)

	require.Equal(t, int64(20), result[1].InternalID, "r2 InternalID preserved")
	require.Equal(t, int64(3), result[1].Troops)
	require.Equal(t, "player1", result[1].OwnerID)

	// Unchanged region passes through with InternalID intact.
	require.Equal(t, int64(30), result[2].InternalID, "r3 InternalID preserved")
	require.Equal(t, int64(2), result[2].Troops)
	require.Equal(t, "player1", result[2].OwnerID)
}

func TestMergeRegions_NoUpdatesPreservesAll(t *testing.T) {
	t.Parallel()

	base := []snapshot.RegionState{
		{InternalID: 10, ID: "r1", OwnerID: "player1", Troops: 5},
	}

	result := attack.MergeRegions(base, nil)

	require.Len(t, result, 1)
	require.Equal(t, int64(10), result[0].InternalID)
	require.Equal(t, "r1", result[0].ID)
}
