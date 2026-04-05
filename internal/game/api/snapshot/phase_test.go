package snapshot_test

import (
	"encoding/json"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/stretchr/testify/require"
)

func TestPhase_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		phase snapshot.Phase
	}{
		{
			name: "cards phase with empty state",
			phase: snapshot.Phase{
				Type:  snapshot.PhaseCards,
				State: snapshot.EmptyPhaseState{},
			},
		},
		{
			name: "attack phase with empty state",
			phase: snapshot.Phase{
				Type:  snapshot.PhaseAttack,
				State: snapshot.EmptyPhaseState{},
			},
		},
		{
			name: "reinforce phase with empty state",
			phase: snapshot.Phase{
				Type:  snapshot.PhaseReinforce,
				State: snapshot.EmptyPhaseState{},
			},
		},
		{
			name: "deploy phase with deployable troops",
			phase: snapshot.Phase{
				Type:  snapshot.PhaseDeploy,
				State: snapshot.DeployPhaseState{DeployableTroops: 7},
			},
		},
		{
			name: "conquer phase with region data",
			phase: snapshot.Phase{
				Type: snapshot.PhaseConquer,
				State: snapshot.ConquerPhaseState{
					AttackingRegionID: "alaska",
					DefendingRegionID: "kamchatka",
					MinTroopsToMove:   2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.phase)
			require.NoError(t, err)

			var got snapshot.Phase
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)

			require.Equal(t, tt.phase.Type, got.Type)
			require.Equal(t, tt.phase.State, got.State)
		})
	}
}

func TestPhase_WireCompatibility(t *testing.T) {
	t.Parallel()

	// The sealed Phase wrapper must produce identical JSON to an anonymous
	// struct with the same discriminator + state layout. This is the pattern
	// used in converter/public.go today.
	tests := []struct {
		name          string
		anonymousJSON any
		sealedPhase   snapshot.Phase
	}{
		{
			name: "deploy phase",
			anonymousJSON: struct {
				Type  snapshot.PhaseType `json:"type"`
				State any                `json:"state"`
			}{
				Type:  snapshot.PhaseDeploy,
				State: snapshot.DeployPhaseState{DeployableTroops: 7},
			},
			sealedPhase: snapshot.Phase{
				Type:  snapshot.PhaseDeploy,
				State: snapshot.DeployPhaseState{DeployableTroops: 7},
			},
		},
		{
			name: "conquer phase",
			anonymousJSON: struct {
				Type  snapshot.PhaseType `json:"type"`
				State any                `json:"state"`
			}{
				Type: snapshot.PhaseConquer,
				State: snapshot.ConquerPhaseState{
					AttackingRegionID: "alaska",
					DefendingRegionID: "kamchatka",
					MinTroopsToMove:   2,
				},
			},
			sealedPhase: snapshot.Phase{
				Type: snapshot.PhaseConquer,
				State: snapshot.ConquerPhaseState{
					AttackingRegionID: "alaska",
					DefendingRegionID: "kamchatka",
					MinTroopsToMove:   2,
				},
			},
		},
		{
			name: "attack phase with empty state",
			anonymousJSON: struct {
				Type  snapshot.PhaseType `json:"type"`
				State any                `json:"state"`
			}{
				Type:  snapshot.PhaseAttack,
				State: snapshot.EmptyPhaseState{},
			},
			sealedPhase: snapshot.Phase{
				Type:  snapshot.PhaseAttack,
				State: snapshot.EmptyPhaseState{},
			},
		},
		{
			name: "cards phase with empty state",
			anonymousJSON: struct {
				Type  snapshot.PhaseType `json:"type"`
				State any                `json:"state"`
			}{
				Type:  snapshot.PhaseCards,
				State: snapshot.EmptyPhaseState{},
			},
			sealedPhase: snapshot.Phase{
				Type:  snapshot.PhaseCards,
				State: snapshot.EmptyPhaseState{},
			},
		},
		{
			name: "reinforce phase with empty state",
			anonymousJSON: struct {
				Type  snapshot.PhaseType `json:"type"`
				State any                `json:"state"`
			}{
				Type:  snapshot.PhaseReinforce,
				State: snapshot.EmptyPhaseState{},
			},
			sealedPhase: snapshot.Phase{
				Type:  snapshot.PhaseReinforce,
				State: snapshot.EmptyPhaseState{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			existingJSON, err := json.Marshal(tt.anonymousJSON)
			require.NoError(t, err)

			sealedJSON, err := json.Marshal(tt.sealedPhase)
			require.NoError(t, err)

			require.JSONEq(t, string(existingJSON), string(sealedJSON),
				"wire format mismatch: existing=%s sealed=%s", existingJSON, sealedJSON)
		})
	}
}

func TestPhase_UnknownType(t *testing.T) {
	t.Parallel()

	data := []byte(`{"type":"unknown_phase","state":{}}`)

	var got snapshot.Phase
	err := json.Unmarshal(data, &got)
	require.Error(t, err)
	require.ErrorContains(t, err, `unknown phase type: "unknown_phase"`)
}

func TestPhase_MalformedJSON(t *testing.T) {
	t.Parallel()

	t.Run("truncated JSON object", func(t *testing.T) {
		t.Parallel()

		var got snapshot.Phase
		err := json.Unmarshal([]byte(`{"type":"deploy","state":`), &got)
		require.Error(t, err)
	})

	t.Run("invalid field type in deploy state", func(t *testing.T) {
		t.Parallel()

		var got snapshot.Phase
		err := json.Unmarshal(
			[]byte(`{"type":"deploy","state":{"deployableTroops":"not_a_number"}}`),
			&got,
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "unmarshaling deploy phase state")
	})

	t.Run("invalid field type in conquer state", func(t *testing.T) {
		t.Parallel()

		var got snapshot.Phase
		err := json.Unmarshal(
			[]byte(`{"type":"conquer","state":{"attackingRegionId":123}}`),
			&got,
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "unmarshaling conquer phase state")
	})
}
