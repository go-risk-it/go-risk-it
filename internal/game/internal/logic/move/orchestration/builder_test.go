package orchestration_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/stretchr/testify/require"
)

// --- Fixture builders ---

func basePrevState() *snapshot.CachedGameState {
	return &snapshot.CachedGameState{
		Turn:            5,
		ConqueredInTurn: false,
		PublicSnapshot: &snapshot.GameSnapshot{
			Game: snapshot.GameMeta{
				ID:   42,
				Turn: 5,
			},
			Phase: snapshot.Phase{
				Type:  snapshot.PhaseAttack,
				State: snapshot.EmptyPhaseState{},
			},
			Regions: []snapshot.RegionState{
				{ID: "western-europe", OwnerID: "player1", Troops: 3},
				{ID: "eastern-europe", OwnerID: "player2", Troops: 2},
				{ID: "north-africa", OwnerID: "player1", Troops: 5},
				{ID: "brazil", OwnerID: "player3", Troops: 1},
			},
			Players: []snapshot.PlayerState{
				{
					UserID:    "player1",
					Name:      "Alice",
					Index:     0,
					CardCount: 2,
					Status:    snapshot.PlayerAlive,
				},
				{
					UserID:    "player2",
					Name:      "Bob",
					Index:     1,
					CardCount: 1,
					Status:    snapshot.PlayerAlive,
				},
				{
					UserID:    "player3",
					Name:      "Charlie",
					Index:     2,
					CardCount: 0,
					Status:    snapshot.PlayerAlive,
				},
			},
		},
		PrivateSnapshots: map[string]*snapshot.PlayerPrivate{
			"player1": {
				Cards: []snapshot.CardState{
					{ID: 10, Type: snapshot.CardInfantry, Region: "western-europe"},
					{ID: 11, Type: snapshot.CardCavalry, Region: "north-africa"},
				},
				Mission: snapshot.PlayerMission{
					Type:   snapshot.MissionTwentyFourTerritories,
					Detail: snapshot.TwentyFourTerritoriesMission{},
				},
			},
			"player2": {
				Cards: []snapshot.CardState{
					{ID: 20, Type: snapshot.CardArtillery, Region: "eastern-europe"},
				},
				Mission: snapshot.PlayerMission{
					Type:   snapshot.MissionEliminatePlayer,
					Detail: snapshot.EliminatePlayerMission{TargetUserID: "player3"},
				},
			},
			"player3": {
				Cards: []snapshot.CardState{},
				Mission: snapshot.PlayerMission{
					Type: snapshot.MissionTwoContinents,
					Detail: snapshot.TwoContinentsMission{
						Continent1: "europe",
						Continent2: "asia",
					},
				},
			},
		},
	}
}

func emptyMoveEffect() *service.MoveEffect {
	return &service.MoveEffect{
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}
}

func TestBuildNewState_DeployRegionTroopIncrease(t *testing.T) {
	prev := basePrevState()
	prev.PublicSnapshot.Phase = snapshot.Phase{
		Type:  snapshot.PhaseDeploy,
		State: snapshot.DeployPhaseState{DeployableTroops: 5},
	}

	effect := &service.MoveEffect{
		RegionUpdates: []service.RegionUpdate{
			{RegionID: "western-europe", NewOwner: "player1", NewTroops: 6},
		},
		UpdatedPhase: snapshot.DeployPhaseState{DeployableTroops: 2},
	}

	result := orchestration.BuildNewState(
		prev, effect, nil, sqlc.GamePhaseTypeDEPLOY, "",
	)

	require.Equal(t, int64(6), findRegion(t, result, "western-europe").Troops)
	require.Equal(t, "player1", findRegion(t, result, "western-europe").OwnerID)
	// Unchanged regions should be carried forward.
	require.Equal(t, int64(2), findRegion(t, result, "eastern-europe").Troops)
}

func TestBuildNewState_AttackConquestOwnershipChange(t *testing.T) {
	prev := basePrevState()

	effect := &service.MoveEffect{
		RegionUpdates: []service.RegionUpdate{
			{RegionID: "western-europe", NewOwner: "player1", NewTroops: 1},
			{RegionID: "eastern-europe", NewOwner: "player1", NewTroops: 2},
		},
		UpdatedPhase: snapshot.ConquerPhaseState{
			AttackingRegionID: "western-europe",
			DefendingRegionID: "eastern-europe",
			MinTroopsToMove:   1,
		},
	}

	result := orchestration.BuildNewState(
		prev, effect, nil, sqlc.GamePhaseTypeCONQUER, "",
	)

	// eastern-europe should now belong to player1.
	require.Equal(t, "player1", findRegion(t, result, "eastern-europe").OwnerID)
	require.Equal(t, int64(2), findRegion(t, result, "eastern-europe").Troops)
	// Phase should be conquer with the right state.
	require.Equal(t, snapshot.PhaseConquer, result.PublicSnapshot.Phase.Type)
	cs, ok := result.PublicSnapshot.Phase.State.(snapshot.ConquerPhaseState)
	require.True(t, ok)
	require.Equal(t, "western-europe", cs.AttackingRegionID)
	// ConqueredInTurn should be set true entering CONQUER.
	require.True(t, result.ConqueredInTurn)
}

func TestBuildNewState_CardDeltaApplied(t *testing.T) {
	prev := basePrevState()

	effect := &service.MoveEffect{
		CardDeltas: []service.CardDelta{
			{
				PlayerUserID: "player1",
				Gained:       []snapshot.CardState{{ID: 30, Type: snapshot.CardJolly, Region: ""}},
				Lost:         []int64{10},
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	result := orchestration.BuildNewState(
		prev, effect, nil, sqlc.GamePhaseTypeATTACK, "",
	)

	priv := result.PrivateSnapshots["player1"]
	require.Len(t, priv.Cards, 2) // started with 2, lost 1, gained 1
	// Card 10 should be gone.
	for _, c := range priv.Cards {
		require.NotEqual(t, int64(10), c.ID)
	}
	// Card 30 should be present.
	found := false
	for _, c := range priv.Cards {
		if c.ID == 30 {
			found = true
		}
	}
	require.True(t, found, "card 30 should be present")

	// Public player card count should reflect the change.
	p1 := findPlayer(t, result, "player1")
	require.Equal(t, int64(2), p1.CardCount)
}

func TestBuildNewState_CardDeltaFromAdvanceEffect(t *testing.T) {
	prev := basePrevState()

	advEffect := &service.AdvanceEffect{
		NewPhase:  snapshot.DeployPhaseState{DeployableTroops: 10},
		TurnEnded: true,
		CardDeltas: []service.CardDelta{
			{
				PlayerUserID: "player1",
				Gained: []snapshot.CardState{
					{ID: 50, Type: snapshot.CardArtillery, Region: "brazil"},
				},
			},
		},
	}

	result := orchestration.BuildNewState(
		prev, emptyMoveEffect(), advEffect, sqlc.GamePhaseTypeDEPLOY, "",
	)

	priv := result.PrivateSnapshots["player1"]
	require.Len(t, priv.Cards, 3) // 2 original + 1 gained from advance
	p1 := findPlayer(t, result, "player1")
	require.Equal(t, int64(3), p1.CardCount)
}

func TestBuildNewState_PlayerElimination(t *testing.T) {
	prev := basePrevState()

	// player3 owns only "brazil". Take it from them.
	effect := &service.MoveEffect{
		RegionUpdates: []service.RegionUpdate{
			{RegionID: "brazil", NewOwner: "player1", NewTroops: 2},
		},
		UpdatedPhase: snapshot.ConquerPhaseState{
			AttackingRegionID: "north-africa",
			DefendingRegionID: "brazil",
			MinTroopsToMove:   1,
		},
	}

	result := orchestration.BuildNewState(
		prev, effect, nil, sqlc.GamePhaseTypeCONQUER, "",
	)

	p3 := findPlayer(t, result, "player3")
	require.Equal(t, snapshot.PlayerDead, p3.Status)
	// player1 should still be alive.
	p1 := findPlayer(t, result, "player1")
	require.Equal(t, snapshot.PlayerAlive, p1.Status)
}

func TestBuildNewState_TurnIncrement(t *testing.T) {
	prev := basePrevState()

	advEffect := &service.AdvanceEffect{
		NewPhase:  snapshot.DeployPhaseState{DeployableTroops: 10},
		TurnEnded: true,
	}

	result := orchestration.BuildNewState(
		prev, emptyMoveEffect(), advEffect, sqlc.GamePhaseTypeDEPLOY, "",
	)

	require.Equal(t, int64(6), result.Turn)
	require.Equal(t, int64(6), result.PublicSnapshot.Game.Turn)
}

func TestBuildNewState_Immutability(t *testing.T) {
	prev := basePrevState()

	// Save copies of values we'll check.
	origRegionTroops := prev.PublicSnapshot.Regions[0].Troops
	origPlayer1Cards := len(prev.PrivateSnapshots["player1"].Cards)
	origTurn := prev.Turn
	origCard0ID := prev.PrivateSnapshots["player1"].Cards[0].ID

	effect := &service.MoveEffect{
		RegionUpdates: []service.RegionUpdate{
			{RegionID: "western-europe", NewOwner: "player1", NewTroops: 99},
		},
		CardDeltas: []service.CardDelta{
			{
				PlayerUserID: "player1",
				Gained:       []snapshot.CardState{{ID: 99, Type: snapshot.CardJolly, Region: ""}},
				Lost:         []int64{10},
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	advEffect := &service.AdvanceEffect{
		NewPhase:  snapshot.DeployPhaseState{DeployableTroops: 10},
		TurnEnded: true,
	}

	_ = orchestration.BuildNewState(
		prev, effect, advEffect, sqlc.GamePhaseTypeDEPLOY, "",
	)

	// Verify prev is completely untouched.
	require.Equal(t, origTurn, prev.Turn)
	require.Equal(t, origRegionTroops, prev.PublicSnapshot.Regions[0].Troops)
	require.Len(t, prev.PrivateSnapshots["player1"].Cards, origPlayer1Cards)
	require.Equal(t, origCard0ID, prev.PrivateSnapshots["player1"].Cards[0].ID)
}

func TestBuildNewState_IdentityOnEmptyEffects(t *testing.T) {
	prev := basePrevState()

	result := orchestration.BuildNewState(
		prev, emptyMoveEffect(), nil, sqlc.GamePhaseTypeATTACK, "",
	)

	// State should be semantically identical to prev.
	require.Equal(t, prev.Turn, result.Turn)
	require.Len(t, result.PublicSnapshot.Regions, len(prev.PublicSnapshot.Regions))
	for i, r := range result.PublicSnapshot.Regions {
		require.Equal(t, prev.PublicSnapshot.Regions[i], r)
	}
	require.Equal(t, prev.ConqueredInTurn, result.ConqueredInTurn)
	// But it must be a different pointer.
	require.NotSame(t, prev.PublicSnapshot, result.PublicSnapshot)
}

func TestBuildNewState_ConqueredInTurnLifecycle(t *testing.T) {
	tests := []struct {
		name        string
		prevCIT     bool
		targetPhase sqlc.GamePhaseType
		turnEnded   bool
		wantCIT     bool
	}{
		{
			name:        "set true entering conquer phase",
			prevCIT:     false,
			targetPhase: sqlc.GamePhaseTypeCONQUER,
			turnEnded:   false,
			wantCIT:     true,
		},
		{
			name:        "reset on turn end",
			prevCIT:     true,
			targetPhase: sqlc.GamePhaseTypeDEPLOY,
			turnEnded:   true,
			wantCIT:     false,
		},
		{
			name:        "carry forward in attack phase",
			prevCIT:     true,
			targetPhase: sqlc.GamePhaseTypeATTACK,
			turnEnded:   false,
			wantCIT:     true,
		},
		{
			name:        "carry forward false in reinforce phase",
			prevCIT:     false,
			targetPhase: sqlc.GamePhaseTypeREINFORCE,
			turnEnded:   false,
			wantCIT:     false,
		},
		{
			name:        "set true even if already true",
			prevCIT:     true,
			targetPhase: sqlc.GamePhaseTypeCONQUER,
			turnEnded:   false,
			wantCIT:     true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			prev := basePrevState()
			prev.ConqueredInTurn = testCase.prevCIT

			var advEffect *service.AdvanceEffect
			if testCase.turnEnded {
				advEffect = &service.AdvanceEffect{
					NewPhase:  snapshot.DeployPhaseState{DeployableTroops: 10},
					TurnEnded: true,
				}
			}

			result := orchestration.BuildNewState(
				prev, emptyMoveEffect(), advEffect, testCase.targetPhase, "",
			)

			require.Equal(t, testCase.wantCIT, result.ConqueredInTurn)
		})
	}
}

func TestBuildNewState_PhaseSelection(t *testing.T) {
	t.Run("advance effect phase takes precedence when present", func(t *testing.T) {
		prev := basePrevState()

		effect := &service.MoveEffect{
			UpdatedPhase: snapshot.EmptyPhaseState{},
		}
		advEffect := &service.AdvanceEffect{
			NewPhase: snapshot.DeployPhaseState{DeployableTroops: 10},
		}

		result := orchestration.BuildNewState(
			prev, effect, advEffect, sqlc.GamePhaseTypeDEPLOY, "",
		)

		require.Equal(t, snapshot.PhaseDeploy, result.PublicSnapshot.Phase.Type)
		ds, ok := result.PublicSnapshot.Phase.State.(snapshot.DeployPhaseState)
		require.True(t, ok)
		require.Equal(t, int64(10), ds.DeployableTroops)
	})

	t.Run("move effect phase used when no advance", func(t *testing.T) {
		prev := basePrevState()

		effect := &service.MoveEffect{
			UpdatedPhase: snapshot.ConquerPhaseState{
				AttackingRegionID: "western-europe",
				DefendingRegionID: "eastern-europe",
				MinTroopsToMove:   1,
			},
		}

		result := orchestration.BuildNewState(
			prev, effect, nil, sqlc.GamePhaseTypeCONQUER, "",
		)

		require.Equal(t, snapshot.PhaseConquer, result.PublicSnapshot.Phase.Type)
	})
}

func TestBuildNewState_GameOver(t *testing.T) {
	prev := basePrevState()

	result := orchestration.BuildNewState(
		prev, emptyMoveEffect(), nil, sqlc.GamePhaseTypeATTACK, "player1",
	)

	require.Equal(t, "player1", result.PublicSnapshot.Game.WinnerUserID)
}

func TestBuildNewState_MissionChange(t *testing.T) {
	prev := basePrevState()

	effect := &service.MoveEffect{
		Missions: []service.MissionChange{
			{
				PlayerUserID: "player2",
				NewMission: snapshot.PlayerMission{
					Type:   snapshot.MissionTwentyFourTerritories,
					Detail: snapshot.TwentyFourTerritoriesMission{},
				},
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	result := orchestration.BuildNewState(
		prev, effect, nil, sqlc.GamePhaseTypeATTACK, "",
	)

	require.Equal(t,
		snapshot.MissionTwentyFourTerritories,
		result.PrivateSnapshots["player2"].Mission.Type,
	)
	// Original should be unchanged.
	require.Equal(t,
		snapshot.MissionEliminatePlayer,
		prev.PrivateSnapshots["player2"].Mission.Type,
	)
}

func TestBuildNewState_PanicsOnUnknownRegion(t *testing.T) {
	prev := basePrevState()

	effect := &service.MoveEffect{
		RegionUpdates: []service.RegionUpdate{
			{RegionID: "atlantis", NewOwner: "player1", NewTroops: 5},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	require.Panics(t, func() {
		orchestration.BuildNewState(prev, effect, nil, sqlc.GamePhaseTypeATTACK, "")
	})
}

func TestBuildNewState_PanicsOnNilPrev(t *testing.T) {
	require.Panics(t, func() {
		orchestration.BuildNewState(
			nil,
			emptyMoveEffect(),
			nil,
			sqlc.GamePhaseTypeATTACK,
			"",
		)
	})
}

func TestBuildNewState_PanicsOnNilMoveEffect(t *testing.T) {
	prev := basePrevState()
	require.Panics(t, func() {
		orchestration.BuildNewState(prev, nil, nil, sqlc.GamePhaseTypeATTACK, "")
	})
}

// --- Edge case tests ---

func TestBuildNewState_MultipleRegionUpdatesConquerScenario(t *testing.T) {
	// Simulate a conquer move that updates 3 regions: attacker loses troops,
	// defender changes owner, and a third region is affected (e.g., reinforcement
	// redistribution in the same effect).
	prev := basePrevState()

	effect := &service.MoveEffect{
		RegionUpdates: []service.RegionUpdate{
			{
				RegionID:  "western-europe",
				NewOwner:  "player1",
				NewTroops: 1,
			}, // attacker spent troops
			{RegionID: "eastern-europe", NewOwner: "player1", NewTroops: 3}, // conquered
			{RegionID: "north-africa", NewOwner: "player1", NewTroops: 2},   // redistributed
		},
		UpdatedPhase: snapshot.ConquerPhaseState{
			AttackingRegionID: "western-europe",
			DefendingRegionID: "eastern-europe",
			MinTroopsToMove:   1,
		},
	}

	result := orchestration.BuildNewState(
		prev, effect, nil, sqlc.GamePhaseTypeCONQUER, "",
	)

	require.Equal(t, int64(1), findRegion(t, result, "western-europe").Troops)
	require.Equal(t, "player1", findRegion(t, result, "eastern-europe").OwnerID)
	require.Equal(t, int64(3), findRegion(t, result, "eastern-europe").Troops)
	require.Equal(t, int64(2), findRegion(t, result, "north-africa").Troops)
	// brazil should be untouched.
	require.Equal(t, "player3", findRegion(t, result, "brazil").OwnerID)
	require.Equal(t, int64(1), findRegion(t, result, "brazil").Troops)
}

func TestBuildNewState_CardTransferMultipleCardsOnElimination(t *testing.T) {
	// When a player is eliminated, the attacker receives all of their cards.
	// This test verifies that multiple cards transfer correctly in a single delta.
	prev := basePrevState()
	// Give player3 multiple cards for the transfer.
	prev.PrivateSnapshots["player3"] = &snapshot.PlayerPrivate{
		Cards: []snapshot.CardState{
			{ID: 40, Type: snapshot.CardInfantry, Region: "brazil"},
			{ID: 41, Type: snapshot.CardCavalry, Region: "eastern-europe"},
			{ID: 42, Type: snapshot.CardArtillery, Region: "western-europe"},
		},
		Mission: snapshot.PlayerMission{
			Type:   snapshot.MissionTwoContinents,
			Detail: snapshot.TwoContinentsMission{Continent1: "europe", Continent2: "asia"},
		},
	}

	effect := &service.MoveEffect{
		RegionUpdates: []service.RegionUpdate{
			{RegionID: "brazil", NewOwner: "player1", NewTroops: 2},
		},
		CardDeltas: []service.CardDelta{
			// player3 loses all cards
			{
				PlayerUserID: "player3",
				Lost:         []int64{40, 41, 42},
			},
			// player1 gains them
			{
				PlayerUserID: "player1",
				Gained: []snapshot.CardState{
					{ID: 40, Type: snapshot.CardInfantry, Region: "brazil"},
					{ID: 41, Type: snapshot.CardCavalry, Region: "eastern-europe"},
					{ID: 42, Type: snapshot.CardArtillery, Region: "western-europe"},
				},
			},
		},
		UpdatedPhase: snapshot.ConquerPhaseState{
			AttackingRegionID: "north-africa",
			DefendingRegionID: "brazil",
			MinTroopsToMove:   1,
		},
	}

	result := orchestration.BuildNewState(
		prev, effect, nil, sqlc.GamePhaseTypeCONQUER, "",
	)

	// player3 should have 0 cards and be dead.
	p3Priv := result.PrivateSnapshots["player3"]
	require.Empty(t, p3Priv.Cards)
	require.Equal(t, snapshot.PlayerDead, findPlayer(t, result, "player3").Status)

	// player1 should have 2 original + 3 transferred = 5 cards.
	p1Priv := result.PrivateSnapshots["player1"]
	require.Len(t, p1Priv.Cards, 5)
	require.Equal(t, int64(5), findPlayer(t, result, "player1").CardCount)

	// Verify all 3 transferred cards are present.
	cardIDs := make(map[int64]bool)
	for _, c := range p1Priv.Cards {
		cardIDs[c.ID] = true
	}

	require.True(t, cardIDs[40], "transferred card 40 should be present")
	require.True(t, cardIDs[41], "transferred card 41 should be present")
	require.True(t, cardIDs[42], "transferred card 42 should be present")
}

func TestBuildNewState_MoveAndAdvanceCardDeltasAppliedInOrder(t *testing.T) {
	// The move effect might remove cards (e.g., cards phase plays them),
	// and the advance effect might grant a card (e.g., end-of-turn conquest reward).
	// They must be applied sequentially: move deltas first, advance deltas second.
	prev := basePrevState()

	moveEffect := &service.MoveEffect{
		CardDeltas: []service.CardDelta{
			{
				// Cards phase: player1 plays all their cards.
				PlayerUserID: "player1",
				Lost:         []int64{10, 11},
			},
		},
		UpdatedPhase: snapshot.DeployPhaseState{DeployableTroops: 8},
	}

	advEffect := &service.AdvanceEffect{
		NewPhase:  snapshot.DeployPhaseState{DeployableTroops: 8},
		TurnEnded: false,
		CardDeltas: []service.CardDelta{
			{
				// Conquest reward: player1 gets a new card.
				PlayerUserID: "player1",
				Gained:       []snapshot.CardState{{ID: 60, Type: snapshot.CardJolly, Region: ""}},
			},
		},
	}

	result := orchestration.BuildNewState(
		prev, moveEffect, advEffect, sqlc.GamePhaseTypeDEPLOY, "",
	)

	p1Priv := result.PrivateSnapshots["player1"]
	// Started with 2 (IDs 10, 11), lost both in move, gained 1 in advance.
	require.Len(t, p1Priv.Cards, 1)
	require.Equal(t, int64(60), p1Priv.Cards[0].ID)
	require.Equal(t, int64(1), findPlayer(t, result, "player1").CardCount)
}

func TestBuildNewState_AdvanceCardDeltaSeesPostMoveState(t *testing.T) {
	// Verify that if the move effect gains a card and the advance effect loses it,
	// the final state reflects the full chain (not just one or the other).
	prev := basePrevState()

	moveEffect := &service.MoveEffect{
		CardDeltas: []service.CardDelta{
			{
				PlayerUserID: "player1",
				Gained: []snapshot.CardState{
					{ID: 70, Type: snapshot.CardInfantry, Region: "brazil"},
				},
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	advEffect := &service.AdvanceEffect{
		NewPhase:  snapshot.EmptyPhaseState{},
		TurnEnded: false,
		CardDeltas: []service.CardDelta{
			{
				// Advance removes the card that was gained by the move.
				PlayerUserID: "player1",
				Lost:         []int64{70},
			},
		},
	}

	result := orchestration.BuildNewState(
		prev, moveEffect, advEffect, sqlc.GamePhaseTypeATTACK, "",
	)

	p1Priv := result.PrivateSnapshots["player1"]
	// Started with 2 cards, gained 1 in move, lost it in advance => 2.
	require.Len(t, p1Priv.Cards, 2)

	for _, c := range p1Priv.Cards {
		require.NotEqual(t, int64(70), c.ID, "card 70 should have been removed by advance delta")
	}
}

func TestBuildNewState_RecomputePlayersAliveDeadMixed(t *testing.T) {
	tests := []struct {
		name          string
		regionUpdates []service.RegionUpdate
		wantStatus    map[string]snapshot.PlayerStatus
		wantCards     map[string]int64
	}{
		{
			name:          "all alive with original ownership",
			regionUpdates: nil,
			wantStatus: map[string]snapshot.PlayerStatus{
				"player1": snapshot.PlayerAlive,
				"player2": snapshot.PlayerAlive,
				"player3": snapshot.PlayerAlive,
			},
			wantCards: map[string]int64{
				"player1": 2,
				"player2": 1,
				"player3": 0,
			},
		},
		{
			name: "one player eliminated",
			regionUpdates: []service.RegionUpdate{
				{RegionID: "brazil", NewOwner: "player1", NewTroops: 2},
			},
			wantStatus: map[string]snapshot.PlayerStatus{
				"player1": snapshot.PlayerAlive,
				"player2": snapshot.PlayerAlive,
				"player3": snapshot.PlayerDead,
			},
			wantCards: map[string]int64{
				"player1": 2,
				"player2": 1,
				"player3": 0,
			},
		},
		{
			name: "two players eliminated leaving one survivor",
			regionUpdates: []service.RegionUpdate{
				{RegionID: "eastern-europe", NewOwner: "player1", NewTroops: 3},
				{RegionID: "brazil", NewOwner: "player1", NewTroops: 2},
			},
			wantStatus: map[string]snapshot.PlayerStatus{
				"player1": snapshot.PlayerAlive,
				"player2": snapshot.PlayerDead,
				"player3": snapshot.PlayerDead,
			},
			wantCards: map[string]int64{
				"player1": 2,
				"player2": 1,
				"player3": 0,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			prev := basePrevState()

			effect := &service.MoveEffect{
				RegionUpdates: testCase.regionUpdates,
				UpdatedPhase:  snapshot.EmptyPhaseState{},
			}

			result := orchestration.BuildNewState(
				prev, effect, nil, sqlc.GamePhaseTypeATTACK, "",
			)

			for userID, wantSt := range testCase.wantStatus {
				p := findPlayer(t, result, userID)
				require.Equal(t, wantSt, p.Status, "player %s status", userID)
			}

			for userID, wantCC := range testCase.wantCards {
				p := findPlayer(t, result, userID)
				require.Equal(t, wantCC, p.CardCount, "player %s card count", userID)
			}
		})
	}
}

func TestBuildNewState_AllPhaseTypeMappings(t *testing.T) {
	tests := []struct {
		sqlcPhase    sqlc.GamePhaseType
		wantSnapshot snapshot.PhaseType
	}{
		{sqlc.GamePhaseTypeCARDS, snapshot.PhaseCards},
		{sqlc.GamePhaseTypeDEPLOY, snapshot.PhaseDeploy},
		{sqlc.GamePhaseTypeATTACK, snapshot.PhaseAttack},
		{sqlc.GamePhaseTypeCONQUER, snapshot.PhaseConquer},
		{sqlc.GamePhaseTypeREINFORCE, snapshot.PhaseReinforce},
	}

	for _, testCase := range tests {
		t.Run(string(testCase.sqlcPhase), func(t *testing.T) {
			prev := basePrevState()

			result := orchestration.BuildNewState(
				prev, emptyMoveEffect(), nil, testCase.sqlcPhase, "",
			)

			require.Equal(t, testCase.wantSnapshot, result.PublicSnapshot.Phase.Type)
		})
	}
}

func TestBuildNewState_PanicsOnUnknownPhaseType(t *testing.T) {
	prev := basePrevState()

	require.Panics(t, func() {
		orchestration.BuildNewState(
			prev, emptyMoveEffect(), nil, "INVALID_PHASE", "",
		)
	})
}

func TestBuildNewState_ImmutabilityPrivateSnapshotMap(t *testing.T) {
	prev := basePrevState()

	// Save deep copies of original private state.
	origP1Cards := make([]snapshot.CardState, len(prev.PrivateSnapshots["player1"].Cards))
	copy(origP1Cards, prev.PrivateSnapshots["player1"].Cards)

	origP2Mission := prev.PrivateSnapshots["player2"].Mission
	origP3Cards := make([]snapshot.CardState, len(prev.PrivateSnapshots["player3"].Cards))
	copy(origP3Cards, prev.PrivateSnapshots["player3"].Cards)

	effect := &service.MoveEffect{
		CardDeltas: []service.CardDelta{
			{
				PlayerUserID: "player1",
				Gained:       []snapshot.CardState{{ID: 99, Type: snapshot.CardJolly, Region: ""}},
				Lost:         []int64{10},
			},
			{
				PlayerUserID: "player3",
				Gained: []snapshot.CardState{
					{ID: 98, Type: snapshot.CardInfantry, Region: "brazil"},
				},
			},
		},
		Missions: []service.MissionChange{
			{
				PlayerUserID: "player2",
				NewMission: snapshot.PlayerMission{
					Type:   snapshot.MissionTwentyFourTerritories,
					Detail: snapshot.TwentyFourTerritoriesMission{},
				},
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	advEffect := &service.AdvanceEffect{
		NewPhase:  snapshot.DeployPhaseState{DeployableTroops: 10},
		TurnEnded: true,
		CardDeltas: []service.CardDelta{
			{
				PlayerUserID: "player1",
				Gained: []snapshot.CardState{
					{ID: 97, Type: snapshot.CardCavalry, Region: "eastern-europe"},
				},
			},
		},
	}

	result := orchestration.BuildNewState(
		prev, effect, advEffect, sqlc.GamePhaseTypeDEPLOY, "",
	)

	// Result should have different map entries.
	require.NotSame(t, prev.PrivateSnapshots["player1"], result.PrivateSnapshots["player1"])
	require.NotSame(t, prev.PrivateSnapshots["player2"], result.PrivateSnapshots["player2"])

	// Prev private state must be completely untouched.
	require.Equal(t, origP1Cards, prev.PrivateSnapshots["player1"].Cards)
	require.Equal(t, origP2Mission, prev.PrivateSnapshots["player2"].Mission)
	require.Equal(t, origP3Cards, prev.PrivateSnapshots["player3"].Cards)

	// Verify result has the right mutations applied.
	require.Len(
		t,
		result.PrivateSnapshots["player1"].Cards,
		3,
	) // 2 - 1 + 1(adv) + 1(move gain) = 3
	require.Equal(
		t,
		snapshot.MissionTwentyFourTerritories,
		result.PrivateSnapshots["player2"].Mission.Type,
	)
}

func TestBuildNewState_MissionChangeMultipleTypes(t *testing.T) {
	tests := []struct {
		name       string
		newMission snapshot.PlayerMission
	}{
		{
			name: "eliminate player mission",
			newMission: snapshot.PlayerMission{
				Type:   snapshot.MissionEliminatePlayer,
				Detail: snapshot.EliminatePlayerMission{TargetUserID: "player1"},
			},
		},
		{
			name: "two continents mission",
			newMission: snapshot.PlayerMission{
				Type: snapshot.MissionTwoContinents,
				Detail: snapshot.TwoContinentsMission{
					Continent1: "africa",
					Continent2: "south-america",
				},
			},
		},
		{
			name: "two continents plus one mission",
			newMission: snapshot.PlayerMission{
				Type: snapshot.MissionTwoContinentsPlusOne,
				Detail: snapshot.TwoContinentsPlusOneMission{
					Continent1: "europe",
					Continent2: "north-america",
				},
			},
		},
		{
			name: "eighteen territories two troops mission",
			newMission: snapshot.PlayerMission{
				Type:   snapshot.MissionEighteenTerritoriesTwoTroops,
				Detail: snapshot.EighteenTerritoriesTwoTroopsMission{},
			},
		},
		{
			name: "twenty four territories mission",
			newMission: snapshot.PlayerMission{
				Type:   snapshot.MissionTwentyFourTerritories,
				Detail: snapshot.TwentyFourTerritoriesMission{},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			prev := basePrevState()

			effect := &service.MoveEffect{
				Missions: []service.MissionChange{
					{
						PlayerUserID: "player2",
						NewMission:   testCase.newMission,
					},
				},
				UpdatedPhase: snapshot.EmptyPhaseState{},
			}

			result := orchestration.BuildNewState(
				prev, effect, nil, sqlc.GamePhaseTypeATTACK, "",
			)

			require.Equal(
				t,
				testCase.newMission.Type,
				result.PrivateSnapshots["player2"].Mission.Type,
			)
			require.Equal(
				t,
				testCase.newMission.Detail,
				result.PrivateSnapshots["player2"].Mission.Detail,
			)
			// Prev should be untouched.
			require.Equal(
				t,
				snapshot.MissionEliminatePlayer,
				prev.PrivateSnapshots["player2"].Mission.Type,
			)
		})
	}
}

func TestBuildNewState_CardDeltaForUnknownPlayerSkipped(t *testing.T) {
	// If a card delta references a player not in the private snapshots map,
	// it should be silently skipped (not panic).
	prev := basePrevState()

	effect := &service.MoveEffect{
		CardDeltas: []service.CardDelta{
			{
				PlayerUserID: "nonexistent-player",
				Gained:       []snapshot.CardState{{ID: 99, Type: snapshot.CardJolly, Region: ""}},
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	// Should not panic.
	result := orchestration.BuildNewState(
		prev, effect, nil, sqlc.GamePhaseTypeATTACK, "",
	)

	// All existing players should be unaffected.
	require.Len(t, result.PrivateSnapshots["player1"].Cards, 2)
	require.Len(t, result.PrivateSnapshots["player2"].Cards, 1)
}

func TestBuildNewState_TurnNotIncrementedWithoutTurnEnded(t *testing.T) {
	prev := basePrevState()

	// Advance effect that does NOT end the turn.
	advEffect := &service.AdvanceEffect{
		NewPhase:  snapshot.EmptyPhaseState{},
		TurnEnded: false,
	}

	result := orchestration.BuildNewState(
		prev, emptyMoveEffect(), advEffect, sqlc.GamePhaseTypeATTACK, "",
	)

	require.Equal(t, int64(5), result.Turn)
	require.Equal(t, int64(5), result.PublicSnapshot.Game.Turn)
}

func TestBuildNewState_GameIDPreserved(t *testing.T) {
	prev := basePrevState()

	result := orchestration.BuildNewState(
		prev, emptyMoveEffect(), nil, sqlc.GamePhaseTypeATTACK, "",
	)

	require.Equal(t, int64(42), result.PublicSnapshot.Game.ID)
}

func TestBuildNewState_PlayerNameAndIndexPreserved(t *testing.T) {
	prev := basePrevState()

	result := orchestration.BuildNewState(
		prev, emptyMoveEffect(), nil, sqlc.GamePhaseTypeATTACK, "",
	)

	for _, origPlayer := range prev.PublicSnapshot.Players {
		resultPlayer := findPlayer(t, result, origPlayer.UserID)
		require.Equal(t, origPlayer.Name, resultPlayer.Name, "player %s name", origPlayer.UserID)
		require.Equal(
			t,
			origPlayer.Index,
			resultPlayer.Index,
			"player %s index",
			origPlayer.UserID,
		)
	}
}

func TestBuildNewState_ConqueredInTurnResetTakesPriorityOverConquerPhase(t *testing.T) {
	// Edge case: TurnEnded=true AND targetPhase=CONQUER. The turn-end reset
	// should take priority (resolveConqueredInTurn checks TurnEnded first).
	prev := basePrevState()
	prev.ConqueredInTurn = true

	advEffect := &service.AdvanceEffect{
		NewPhase:  snapshot.EmptyPhaseState{},
		TurnEnded: true,
	}

	result := orchestration.BuildNewState(
		prev, emptyMoveEffect(), advEffect, sqlc.GamePhaseTypeCONQUER, "",
	)

	// TurnEnded wins: should be false even though targetPhase is CONQUER.
	require.False(t, result.ConqueredInTurn)
}

// --- Helpers ---

func findRegion(
	t *testing.T,
	state *snapshot.CachedGameState,
	regionID string,
) snapshot.RegionState {
	t.Helper()

	for _, r := range state.PublicSnapshot.Regions {
		if r.ID == regionID {
			return r
		}
	}

	t.Fatalf("region %q not found", regionID)

	return snapshot.RegionState{}
}

func findPlayer(
	t *testing.T,
	state *snapshot.CachedGameState,
	userID string,
) snapshot.PlayerState {
	t.Helper()

	for _, p := range state.PublicSnapshot.Players {
		if p.UserID == userID {
			return p
		}
	}

	t.Fatalf("player %q not found", userID)

	return snapshot.PlayerState{}
}
