package converter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

// MissionResolver fetches mission details for a given mission type and ID,
// returning the result as a json.RawMessage envelope ready for WS delivery.
// The caller (signal handler) provides the real implementation backed by
// MissionController; tests provide a stub.
type MissionResolver func(
	ctx context.Context,
	missionType sqlc.GameMissionType,
	missionID int64,
) (json.RawMessage, error)

// PublicMessages holds the pre-serialized WS messages for the public broadcast path.
type PublicMessages struct {
	GameState   json.RawMessage
	BoardState  json.RawMessage
	PlayerState json.RawMessage
}

// PrivateMessages holds the pre-serialized WS messages for a single player's
// private write path.
type PrivateMessages struct {
	CardState    json.RawMessage
	MissionState json.RawMessage
}

// convertPhaseType maps sqlc phase types to the API-layer game.PhaseType.
func convertPhaseType(phaseType sqlc.GamePhaseType) (game.PhaseType, error) {
	switch phaseType {
	case sqlc.GamePhaseTypeDEPLOY:
		return game.Deploy, nil
	case sqlc.GamePhaseTypeATTACK:
		return game.Attack, nil
	case sqlc.GamePhaseTypeCONQUER:
		return game.Conquer, nil
	case sqlc.GamePhaseTypeREINFORCE:
		return game.Reinforce, nil
	case sqlc.GamePhaseTypeCARDS:
		return game.Cards, nil
	default:
		return "", fmt.Errorf("unknown phase type: %s", phaseType)
	}
}
