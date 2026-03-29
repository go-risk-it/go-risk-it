package converter

import (
	"context"
	"fmt"

	game "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
)

// MissionResolver fetches mission details for a given mission type and ID,
// returning the result as a typed DTO ready for serialization at the dispatch
// boundary. The caller (publisher) provides the real implementation backed by
// MissionController; tests provide a stub.
type MissionResolver func(
	ctx context.Context,
	missionType sqlc.GameMissionType,
	missionID int64,
) (any, error)

// PublicMessages holds the typed DTOs for the public broadcast path.
// The publisher serializes these into WS message envelopes at dispatch time.
type PublicMessages struct {
	GameState   any
	BoardState  messaging.BoardState
	PlayerState messaging.PlayersState
}

// PrivateMessages holds the typed DTOs for a single player's private write path.
// The publisher serializes these into WS message envelopes at dispatch time.
type PrivateMessages struct {
	CardState    messaging.CardState
	MissionState any
}

// ConvertPhaseType maps sqlc phase types to the API-layer game.PhaseType.
func ConvertPhaseType(phaseType sqlc.GamePhaseType) (game.PhaseType, error) {
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
