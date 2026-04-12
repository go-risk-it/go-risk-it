package orchestration_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/orchestration"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPersistenceEffect_Deploy(t *testing.T) {
	t.Parallel()

	moveCtx := orchestration.NewMoveContext(1, "user1", sqlc.GamePhaseTypeDEPLOY, nil, nil)

	moveEffect := &moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{
				RegionID:  "region1",
				NewOwner:  "user1",
				NewTroops: 5,
			},
		},
		DeployableDelta: -2,
		UpdatedPhase: snapshot.DeployPhaseState{
			DeployableTroops: 3,
		},
	}

	prevState := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Phase: snapshot.Phase{
				Type: snapshot.PhaseDeploy,
				State: snapshot.DeployPhaseState{
					DeployableTroops: 5,
				},
			},
			Regions: []snapshot.RegionState{
				{InternalID: 10, ID: "region1", OwnerID: "user1", Troops: 3},
			},
			Players: []snapshot.PlayerState{
				{UserID: "user1", Index: 0},
			},
		},
	}

	// advEffect has EmptyPhaseState which is same type as current → no phase transition
	result := orchestration.BuildPersistenceEffect(
		moveCtx,
		moveEffect,
		nil,
		prevState,
		sqlc.GamePhaseTypeDEPLOY,
		false,
	)

	require.NotNil(t, result)
	require.NotNil(t, result.MoveLog)
	require.NotNil(t, result.MoveExecution)

	// MoveLog checks
	assert.Equal(t, int64(1), result.MoveLog.GameID)
	assert.Equal(t, "user1", result.MoveLog.UserID)
	assert.Equal(t, "DEPLOY", result.MoveLog.PhaseType)
	assert.NotNil(t, result.MoveLog.MoveData)
	assert.NotNil(t, result.MoveLog.Result)

	// MoveExecution checks — deploy has region updates + deployable delta
	require.Len(t, result.MoveExecution.RegionTroopUpdates, 1)
	assert.Equal(t, int64(10), result.MoveExecution.RegionTroopUpdates[0].RegionID)
	assert.Equal(t, int64(2), result.MoveExecution.RegionTroopUpdates[0].Delta)

	require.NotNil(t, result.MoveExecution.DeployableDelta)
	assert.Equal(t, int64(-2), result.MoveExecution.DeployableDelta.Delta)

	// No elimination, card draw, phase transition, or game conclusion
	assert.Nil(t, result.Elimination)
	assert.Nil(t, result.CardDraw)
	assert.Nil(t, result.PhaseTransition)
	assert.Nil(t, result.GameConclusion)
}

func TestBuildPersistenceEffect_Attack(t *testing.T) {
	t.Parallel()

	moveCtx := orchestration.NewMoveContext(1, "user1", sqlc.GamePhaseTypeATTACK, nil, nil)

	moveEffect := &moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{
				RegionID:  "attacking",
				NewOwner:  "user1",
				NewTroops: 3,
			},
			{
				RegionID:  "defending",
				NewOwner:  "user2",
				NewTroops: 1,
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	prevState := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Regions: []snapshot.RegionState{
				{InternalID: 10, ID: "attacking", OwnerID: "user1", Troops: 5},
				{InternalID: 20, ID: "defending", OwnerID: "user2", Troops: 3},
			},
		},
	}

	result := orchestration.BuildPersistenceEffect(
		moveCtx,
		moveEffect,
		nil,
		prevState,
		sqlc.GamePhaseTypeDEPLOY,
		false,
	)

	require.NotNil(t, result)
	require.NotNil(t, result.MoveLog)
	require.NotNil(t, result.MoveExecution)

	// Attack produces region updates only (troop deltas)
	require.Len(t, result.MoveExecution.RegionTroopUpdates, 2)
	assert.Equal(t, int64(10), result.MoveExecution.RegionTroopUpdates[0].RegionID)
	assert.Equal(t, int64(-2), result.MoveExecution.RegionTroopUpdates[0].Delta)
	assert.Equal(t, int64(20), result.MoveExecution.RegionTroopUpdates[1].RegionID)
	assert.Equal(t, int64(-2), result.MoveExecution.RegionTroopUpdates[1].Delta)

	// No ownership changes in attack (only troops)
	assert.Nil(t, result.MoveExecution.OwnershipChanges)
	assert.Nil(t, result.MoveExecution.DeployableDelta)
	assert.Nil(t, result.Elimination)
	assert.Nil(t, result.CardDraw)
}

func TestBuildPersistenceEffect_Conquer(t *testing.T) {
	t.Parallel()

	moveCtx := orchestration.NewMoveContext(1, "user1", sqlc.GamePhaseTypeCONQUER, nil, nil)

	moveEffect := &moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{
				RegionID:  "source",
				NewOwner:  "user1",
				NewTroops: 2,
			},
			{
				RegionID:  "target",
				NewOwner:  "user1", // ownership changed
				NewTroops: 3,
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	prevState := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Regions: []snapshot.RegionState{
				{InternalID: 10, ID: "source", OwnerID: "user1", Troops: 5},
				{InternalID: 20, ID: "target", OwnerID: "user2", Troops: 0}, // conquered
			},
		},
	}

	result := orchestration.BuildPersistenceEffect(
		moveCtx,
		moveEffect,
		nil,
		prevState,
		sqlc.GamePhaseTypeDEPLOY,
		false,
	)

	require.NotNil(t, result)
	require.NotNil(t, result.MoveExecution)

	// Conquer produces region updates + ownership change
	require.Len(t, result.MoveExecution.RegionTroopUpdates, 2)

	require.Len(t, result.MoveExecution.OwnershipChanges, 1)
	assert.Equal(t, int64(20), result.MoveExecution.OwnershipChanges[0].RegionID)
	assert.Equal(t, "user1", result.MoveExecution.OwnershipChanges[0].NewOwnerUserID)

	// No elimination in this test
	assert.Nil(t, result.Elimination)
}

func TestBuildPersistenceEffect_ConquerWithElimination(t *testing.T) {
	t.Parallel()

	moveCtx := orchestration.NewMoveContext(1, "user1", sqlc.GamePhaseTypeCONQUER, nil, nil)

	moveEffect := &moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{RegionID: "source", NewOwner: "user1", NewTroops: 2},
			{RegionID: "target", NewOwner: "user1", NewTroops: 3},
		},
		EliminatedUserID: "user2", // elimination occurred
		UpdatedPhase:     snapshot.EmptyPhaseState{},
	}

	prevState := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Regions: []snapshot.RegionState{
				{InternalID: 10, ID: "source", OwnerID: "user1", Troops: 5},
				{InternalID: 20, ID: "target", OwnerID: "user2", Troops: 0},
			},
		},
	}

	result := orchestration.BuildPersistenceEffect(
		moveCtx,
		moveEffect,
		nil,
		prevState,
		sqlc.GamePhaseTypeDEPLOY,
		false,
	)

	require.NotNil(t, result)

	// Elimination cascade triggered
	require.NotNil(t, result.Elimination)
	assert.Equal(t, "user2", result.Elimination.EliminatedUserID)
	assert.Equal(t, "user1", result.Elimination.ConquerorUserID)
}

func TestBuildPersistenceEffect_Reinforce(t *testing.T) {
	t.Parallel()

	moveCtx := orchestration.NewMoveContext(1, "user1", sqlc.GamePhaseTypeREINFORCE, nil, nil)

	moveEffect := &moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{RegionID: "source", NewOwner: "user1", NewTroops: 2},
			{RegionID: "target", NewOwner: "user1", NewTroops: 5},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	prevState := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Regions: []snapshot.RegionState{
				{InternalID: 10, ID: "source", OwnerID: "user1", Troops: 5},
				{InternalID: 20, ID: "target", OwnerID: "user1", Troops: 2},
			},
		},
	}

	result := orchestration.BuildPersistenceEffect(
		moveCtx,
		moveEffect,
		nil,
		prevState,
		sqlc.GamePhaseTypeDEPLOY,
		false,
	)

	require.NotNil(t, result)
	require.NotNil(t, result.MoveExecution)

	// Reinforce produces region updates only (no card draw in this case)
	require.Len(t, result.MoveExecution.RegionTroopUpdates, 2)

	assert.Nil(t, result.CardDraw)
}

func TestBuildPersistenceEffect_ReinforceWithCardDraw(t *testing.T) {
	t.Parallel()

	moveCtx := orchestration.NewMoveContext(1, "user1", sqlc.GamePhaseTypeREINFORCE, nil, nil)

	moveEffect := &moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{RegionID: "source", NewOwner: "user1", NewTroops: 2},
			{RegionID: "target", NewOwner: "user1", NewTroops: 5},
		},
		CardDeltas: []moveservice.CardDelta{
			{
				PlayerUserID: "user1",
				Gained: []snapshot.CardState{
					{ID: 100, Type: snapshot.CardInfantry, Region: "region1"},
				},
			},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	prevState := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Regions: []snapshot.RegionState{
				{InternalID: 10, ID: "source", OwnerID: "user1", Troops: 5},
				{InternalID: 20, ID: "target", OwnerID: "user1", Troops: 2},
			},
		},
	}

	result := orchestration.BuildPersistenceEffect(
		moveCtx,
		moveEffect,
		nil,
		prevState,
		sqlc.GamePhaseTypeDEPLOY,
		false,
	)

	require.NotNil(t, result)

	// Card draw present
	require.NotNil(t, result.CardDraw)
	assert.Equal(t, int64(100), result.CardDraw.CardID)
	assert.Equal(t, "user1", result.CardDraw.UserID)
}

func TestBuildPersistenceEffect_Cards(t *testing.T) {
	t.Parallel()

	moveCtx := orchestration.NewMoveContext(1, "user1", sqlc.GamePhaseTypeCARDS, nil, nil)

	moveEffect := &moveservice.MoveEffect{
		CardDeltas: []moveservice.CardDelta{
			{
				PlayerUserID: "user1",
				Lost:         []int64{10, 20, 30},
			},
		},
		RegionUpdates: []moveservice.RegionUpdate{
			{RegionID: "region1", NewOwner: "user1", NewTroops: 5},
		},
		UpdatedPhase: snapshot.EmptyPhaseState{},
	}

	prevState := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Regions: []snapshot.RegionState{
				{InternalID: 100, ID: "region1", OwnerID: "user1", Troops: 3},
			},
		},
	}

	result := orchestration.BuildPersistenceEffect(
		moveCtx,
		moveEffect,
		nil,
		prevState,
		sqlc.GamePhaseTypeDEPLOY,
		false,
	)

	require.NotNil(t, result)
	require.NotNil(t, result.MoveExecution)

	// Cards move produces card unlinks + region bonuses
	require.Len(t, result.MoveExecution.CardUnlinks, 3)
	assert.Contains(t, result.MoveExecution.CardUnlinks, int64(10))
	assert.Contains(t, result.MoveExecution.CardUnlinks, int64(20))
	assert.Contains(t, result.MoveExecution.CardUnlinks, int64(30))

	// Region bonus from card region match
	require.Len(t, result.MoveExecution.RegionTroopUpdates, 1)
	assert.Equal(t, int64(100), result.MoveExecution.RegionTroopUpdates[0].RegionID)
	assert.Equal(t, int64(2), result.MoveExecution.RegionTroopUpdates[0].Delta)
}

func TestBuildPersistenceEffect_VoluntaryAdvance(t *testing.T) {
	t.Parallel()

	moveCtx := orchestration.NewMoveContext(1, "user1", sqlc.GamePhaseTypeATTACK, nil, nil)

	advEffect := &moveservice.AdvanceEffect{
		NewPhase: snapshot.DeployPhaseState{
			DeployableTroops: 5,
		},
		TurnEnded: false,
	}

	prevState := &snapshot.CachedGameState{
		Turn: 1,
		PublicSnapshot: &snapshot.GameSnapshot{
			Phase: snapshot.Phase{
				Type:  snapshot.PhaseAttack,
				State: snapshot.EmptyPhaseState{},
			},
			Players: []snapshot.PlayerState{
				{UserID: "user1", Index: 0},
			},
		},
	}

	result := orchestration.BuildPersistenceEffect(
		moveCtx,
		nil,
		advEffect,
		prevState,
		sqlc.GamePhaseTypeDEPLOY,
		false,
	)

	require.NotNil(t, result)

	// Voluntary advance — no move log
	assert.Nil(t, result.MoveLog)
	assert.Nil(t, result.MoveExecution)

	// Phase transition present
	require.NotNil(t, result.PhaseTransition)
	assert.Equal(t, int64(1), result.PhaseTransition.Turn)
	assert.Equal(t, "DEPLOY", result.PhaseTransition.PhaseType)
	require.Len(t, result.PhaseTransition.Players, 1)
	assert.Equal(t, "user1", result.PhaseTransition.Players[0].UserID)
}

func TestBuildPersistenceEffect_GameOver(t *testing.T) {
	t.Parallel()

	moveCtx := orchestration.NewMoveContext(1, "user1", sqlc.GamePhaseTypeDEPLOY, nil, nil)

	moveEffect := &moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{RegionID: "region1", NewOwner: "user1", NewTroops: 5},
		},
		UpdatedPhase: snapshot.DeployPhaseState{
			DeployableTroops: 0,
		},
	}

	prevState := &snapshot.CachedGameState{
		PublicSnapshot: &snapshot.GameSnapshot{
			Regions: []snapshot.RegionState{
				{InternalID: 10, ID: "region1", OwnerID: "user1", Troops: 3},
			},
		},
	}

	result := orchestration.BuildPersistenceEffect(
		moveCtx,
		moveEffect,
		nil,
		prevState,
		sqlc.GamePhaseTypeDEPLOY,
		true,
	)

	require.NotNil(t, result)

	// Game conclusion populated when gameOver=true
	require.NotNil(t, result.GameConclusion)
	assert.Equal(t, "user1", result.GameConclusion.WinnerUserID)
}

func TestBuildPersistenceEffect_PhaseTransition(t *testing.T) {
	t.Parallel()

	moveCtx := orchestration.NewMoveContext(1, "user1", sqlc.GamePhaseTypeDEPLOY, nil, nil)

	moveEffect := &moveservice.MoveEffect{
		RegionUpdates: []moveservice.RegionUpdate{
			{RegionID: "region1", NewOwner: "user1", NewTroops: 5},
		},
		UpdatedPhase: snapshot.DeployPhaseState{
			DeployableTroops: 0,
		},
	}

	advEffect := &moveservice.AdvanceEffect{
		NewPhase: snapshot.ConquerPhaseState{
			AttackingRegionID: "source",
			DefendingRegionID: "target",
			MinTroopsToMove:   1,
		},
		TurnEnded: true,
	}

	prevState := &snapshot.CachedGameState{
		Turn: 1,
		PublicSnapshot: &snapshot.GameSnapshot{
			Phase: snapshot.Phase{
				Type: snapshot.PhaseDeploy,
				State: snapshot.DeployPhaseState{
					DeployableTroops: 5,
				},
			},
			Regions: []snapshot.RegionState{
				{InternalID: 10, ID: "region1", OwnerID: "user1", Troops: 3},
				{InternalID: 20, ID: "source", OwnerID: "user1", Troops: 5},
				{InternalID: 30, ID: "target", OwnerID: "user2", Troops: 0},
			},
			Players: []snapshot.PlayerState{
				{UserID: "user1", Index: 0},
				{UserID: "user2", Index: 1},
			},
		},
	}

	result := orchestration.BuildPersistenceEffect(
		moveCtx,
		moveEffect,
		advEffect,
		prevState,
		sqlc.GamePhaseTypeCONQUER,
		false,
	)

	require.NotNil(t, result)

	// Phase transition present when advEffect.NewPhase type differs from currentPhase
	require.NotNil(t, result.PhaseTransition)
	assert.Equal(t, int64(2), result.PhaseTransition.Turn) // turn ended
	assert.Equal(t, "CONQUER", result.PhaseTransition.PhaseType)
	require.Len(t, result.PhaseTransition.Players, 2)

	// ConquerData populated
	require.NotNil(t, result.PhaseTransition.ConquerData)
	assert.Equal(t, "source", result.PhaseTransition.ConquerData.SourceRegionName)
	assert.Equal(t, "target", result.PhaseTransition.ConquerData.TargetRegionName)
	assert.Equal(t, int64(1), result.PhaseTransition.ConquerData.MinTroops)
}
