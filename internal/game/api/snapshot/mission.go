package snapshot

import (
	"encoding/json"
	"fmt"
)

// MissionDetail is a sealed interface for mission-specific detail payloads.
// The unexported marker method prevents implementations outside this package.
type MissionDetail interface {
	isMissionDetail() // sealed marker
}

// Compile-time assertions: each variant implements MissionDetail.
var (
	_ MissionDetail = TwoContinentsMission{}
	_ MissionDetail = TwoContinentsPlusOneMission{}
	_ MissionDetail = EighteenTerritoriesTwoTroopsMission{}
	_ MissionDetail = TwentyFourTerritoriesMission{}
	_ MissionDetail = EliminatePlayerMission{}
)

// TwoContinentsMission requires conquering two specific continents.
type TwoContinentsMission struct {
	Continent1 string `json:"continent1"`
	Continent2 string `json:"continent2"`
}

func (TwoContinentsMission) isMissionDetail() {}

// TwoContinentsPlusOneMission requires conquering two specific continents plus
// one additional continent of the player's choice.
type TwoContinentsPlusOneMission struct {
	Continent1 string `json:"continent1"`
	Continent2 string `json:"continent2"`
}

func (TwoContinentsPlusOneMission) isMissionDetail() {}

// EighteenTerritoriesTwoTroopsMission requires conquering 18 territories with
// at least 2 troops each.
type EighteenTerritoriesTwoTroopsMission struct{}

func (EighteenTerritoriesTwoTroopsMission) isMissionDetail() {}

// TwentyFourTerritoriesMission requires conquering 24 territories.
type TwentyFourTerritoriesMission struct{}

func (TwentyFourTerritoriesMission) isMissionDetail() {}

// EliminatePlayerMission requires eliminating a specific player.
type EliminatePlayerMission struct {
	TargetUserID string `json:"targetUserId"`
}

func (EliminatePlayerMission) isMissionDetail() {}

// PlayerMission is the discriminated wrapper for MissionDetail. It carries the
// type discriminator and a runtime-polymorphic Detail field.
type PlayerMission struct {
	Type   MissionType   `json:"type"`
	Detail MissionDetail `json:"details"`
}

// MarshalJSON serializes the PlayerMission with its type discriminator and
// variant-specific details.
func (m PlayerMission) MarshalJSON() ([]byte, error) {
	detailBytes, err := json.Marshal(m.Detail)
	if err != nil {
		return nil, fmt.Errorf("marshaling mission detail: %w", err)
	}

	result, err := json.Marshal(struct {
		Type   MissionType     `json:"type"`
		Detail json.RawMessage `json:"details"`
	}{
		Type:   m.Type,
		Detail: detailBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling player mission: %w", err)
	}

	return result, nil
}

// UnmarshalJSON deserializes a PlayerMission by first reading the type
// discriminator, then routing the details payload to the correct concrete type.
func (m *PlayerMission) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type   MissionType     `json:"type"`
		Detail json.RawMessage `json:"details"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshaling mission wrapper: %w", err)
	}

	m.Type = raw.Type

	return m.unmarshalDetail(raw.Type, raw.Detail)
}

// unmarshalDetail unmarshals the mission detail based on type.
//
//nolint:cyclop // Type discriminator pattern requires one case per type
func (m *PlayerMission) unmarshalDetail(
	missionType MissionType,
	data []byte,
) error {
	switch missionType {
	case MissionTwoContinents:
		var detail TwoContinentsMission
		if err := json.Unmarshal(data, &detail); err != nil {
			return fmt.Errorf("unmarshaling two continents mission: %w", err)
		}
		m.Detail = detail
	case MissionTwoContinentsPlusOne:
		var detail TwoContinentsPlusOneMission
		if err := json.Unmarshal(data, &detail); err != nil {
			return fmt.Errorf("unmarshaling two continents plus one mission: %w", err)
		}
		m.Detail = detail
	case MissionEighteenTerritoriesTwoTroops:
		var detail EighteenTerritoriesTwoTroopsMission
		if err := json.Unmarshal(data, &detail); err != nil {
			return fmt.Errorf("unmarshaling eighteen territories mission: %w", err)
		}
		m.Detail = detail
	case MissionTwentyFourTerritories:
		var detail TwentyFourTerritoriesMission
		if err := json.Unmarshal(data, &detail); err != nil {
			return fmt.Errorf("unmarshaling twenty four territories mission: %w", err)
		}
		m.Detail = detail
	case MissionEliminatePlayer:
		var detail EliminatePlayerMission
		if err := json.Unmarshal(data, &detail); err != nil {
			return fmt.Errorf("unmarshaling eliminate player mission: %w", err)
		}
		m.Detail = detail
	default:
		return fmt.Errorf("unknown mission type: %q", missionType)
	}

	return nil
}
