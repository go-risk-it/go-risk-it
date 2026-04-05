package snapshot

import (
	"encoding/json"
	"fmt"
)

// PhaseState is a sealed interface for phase-specific state payloads.
// The unexported marker method prevents implementations outside this package.
type PhaseState interface {
	isPhaseState() // sealed marker
}

// Compile-time assertions: each variant implements PhaseState.
var (
	_ PhaseState = EmptyPhaseState{}
	_ PhaseState = DeployPhaseState{}
	_ PhaseState = ConquerPhaseState{}
)

// EmptyPhaseState is used for phases with no additional state (cards, attack,
// reinforce).
type EmptyPhaseState struct{}

func (EmptyPhaseState) isPhaseState() {}

// DeployPhaseState holds the number of troops available to deploy.
type DeployPhaseState struct {
	DeployableTroops int64 `json:"deployableTroops"`
}

func (DeployPhaseState) isPhaseState() {}

// ConquerPhaseState holds the region pair and minimum troop movement for a
// conquest.
type ConquerPhaseState struct {
	AttackingRegionID string `json:"attackingRegionId"`
	DefendingRegionID string `json:"defendingRegionId"`
	MinTroopsToMove   int64  `json:"minTroopsToMove"`
}

func (ConquerPhaseState) isPhaseState() {}

// Phase is the discriminated wrapper for PhaseState. It carries the type
// discriminator and a runtime-polymorphic State field.
type Phase struct {
	Type  PhaseType  `json:"type"`
	State PhaseState `json:"state"`
}

// MarshalJSON serializes the Phase with its type discriminator and
// variant-specific state.
func (p Phase) MarshalJSON() ([]byte, error) {
	stateBytes, err := json.Marshal(p.State)
	if err != nil {
		return nil, fmt.Errorf("marshaling phase state: %w", err)
	}

	return json.Marshal(struct {
		Type  PhaseType       `json:"type"`
		State json.RawMessage `json:"state"`
	}{
		Type:  p.Type,
		State: stateBytes,
	})
}

// UnmarshalJSON deserializes a Phase by first reading the type discriminator,
// then routing the state payload to the correct concrete type.
func (p *Phase) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type  PhaseType       `json:"type"`
		State json.RawMessage `json:"state"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshaling phase wrapper: %w", err)
	}

	p.Type = raw.Type

	switch raw.Type {
	case PhaseCards, PhaseAttack, PhaseReinforce:
		var s EmptyPhaseState
		if err := json.Unmarshal(raw.State, &s); err != nil {
			return fmt.Errorf("unmarshaling empty phase state: %w", err)
		}

		p.State = s
	case PhaseDeploy:
		var s DeployPhaseState
		if err := json.Unmarshal(raw.State, &s); err != nil {
			return fmt.Errorf("unmarshaling deploy phase state: %w", err)
		}

		p.State = s
	case PhaseConquer:
		var s ConquerPhaseState
		if err := json.Unmarshal(raw.State, &s); err != nil {
			return fmt.Errorf("unmarshaling conquer phase state: %w", err)
		}

		p.State = s
	default:
		return fmt.Errorf("unknown phase type: %q", raw.Type)
	}

	return nil
}
