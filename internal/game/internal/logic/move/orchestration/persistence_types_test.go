package orchestration_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/orchestration"
	"github.com/stretchr/testify/assert"
)

// TestPersistenceEffect_ZeroValue verifies that a zero-value PersistenceEffect
// has all groups nil, establishing the no-op Persist expectation.
func TestPersistenceEffect_ZeroValue(t *testing.T) {
	var effect orchestration.PersistenceEffect

	assert.Nil(t, effect.MoveLog, "MoveLog should be nil")
	assert.Nil(t, effect.MoveExecution, "MoveExecution should be nil")
	assert.Nil(t, effect.Elimination, "Elimination should be nil")
	assert.Nil(t, effect.CardDraw, "CardDraw should be nil")
	assert.Nil(t, effect.PhaseTransition, "PhaseTransition should be nil")
	assert.Nil(t, effect.GameConclusion, "GameConclusion should be nil")
}

// TestMoveLogEntry_HasPhaseType verifies that MoveLogEntry carries the phase
// type as a direct string field (not a subquery reference).
func TestMoveLogEntry_HasPhaseType(t *testing.T) {
	entry := orchestration.MoveLogEntry{
		GameID:    1,
		UserID:    "user-123",
		PhaseType: "ATTACK",
		MoveData:  []byte(`{"source":"A","target":"B"}`),
		Result:    []byte(`{"outcome":"success"}`),
	}

	assert.Equal(t, "ATTACK", entry.PhaseType, "PhaseType should be a direct field")
	assert.Equal(t, int64(1), entry.GameID)
	assert.Equal(t, "user-123", entry.UserID)
	assert.NotEmpty(t, entry.MoveData)
	assert.NotEmpty(t, entry.Result)
}

// TestPhaseTransition_OptionalSubfields verifies that PhaseTransition contains
// optional ConquerData and DeployData sub-fields.
func TestPhaseTransition_OptionalSubfields(t *testing.T) {
	// Base transition with no optional data.
	base := orchestration.PhaseTransition{
		Turn:      1,
		PhaseType: "ATTACK",
		Players: []orchestration.PlayerRef{
			{UserID: "user-1", InternalID: 10},
			{UserID: "user-2", InternalID: 20},
		},
		ConquerData: nil,
		DeployData:  nil,
	}

	assert.Equal(t, int64(1), base.Turn)
	assert.Equal(t, "ATTACK", base.PhaseType)
	assert.Len(t, base.Players, 2)
	assert.Nil(t, base.ConquerData, "ConquerData is optional")
	assert.Nil(t, base.DeployData, "DeployData is optional")

	// Transition with ConquerData present.
	withConquer := orchestration.PhaseTransition{
		Turn:      2,
		PhaseType: "CONQUER",
		Players: []orchestration.PlayerRef{
			{UserID: "user-1", InternalID: 10},
		},
		ConquerData: &orchestration.ConquerData{
			SourceRegionName: "alaska",
			TargetRegionName: "kamchatka",
			MinTroops:        3,
		},
		DeployData: nil,
	}

	assert.NotNil(t, withConquer.ConquerData, "ConquerData present in CONQUER phase")
	assert.Equal(t, "alaska", withConquer.ConquerData.SourceRegionName)
	assert.Equal(t, "kamchatka", withConquer.ConquerData.TargetRegionName)
	assert.Equal(t, int64(3), withConquer.ConquerData.MinTroops)
	assert.Nil(t, withConquer.DeployData)

	// Transition with DeployData present.
	withDeploy := orchestration.PhaseTransition{
		Turn:      3,
		PhaseType: "DEPLOY",
		Players: []orchestration.PlayerRef{
			{UserID: "user-1", InternalID: 10},
		},
		ConquerData: nil,
		DeployData: &orchestration.DeployData{
			DeployableTroops: 7,
		},
	}

	assert.Nil(t, withDeploy.ConquerData)
	assert.NotNil(t, withDeploy.DeployData, "DeployData present in DEPLOY phase")
	assert.Equal(t, int64(7), withDeploy.DeployData.DeployableTroops)
}
