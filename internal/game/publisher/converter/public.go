package converter

import (
	"errors"
	"fmt"
	"slices"

	game "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/snapshot"
)

// ConvertPublicSnapshot transforms a PublicSnapshot into typed DTOs for the
// broadcast path: gameState, boardState, playerState.
//
// connectedPlayers is the list of user IDs currently connected via WebSocket,
// used to enrich each player's ConnectionStatus field.
func ConvertPublicSnapshot(
	snap *snapshot.PublicSnapshot,
	connectedPlayers []string,
) (*PublicMessages, error) {
	gameState, err := buildGameState(snap)
	if err != nil {
		return nil, fmt.Errorf("building game state: %w", err)
	}

	return &PublicMessages{
		GameState:   gameState,
		BoardState:  buildBoardState(snap.Board),
		PlayerState: buildPlayerState(snap.Players, connectedPlayers),
	}, nil
}

// buildGameState produces the GameState DTO.
// The phase-discriminated generic GameState[T] from messaging cannot be
// used here because Go generics require compile-time type parameters.
// Instead we build the same JSON structure using a local untyped struct,
// which produces byte-identical output when serialized.
func buildGameState(snap *snapshot.PublicSnapshot) (any, error) {
	phaseType, err := ConvertPhaseType(snap.Phase.Type)
	if err != nil {
		return nil, err
	}

	phaseState, err := buildPhaseState(snap.Phase)
	if err != nil {
		return nil, fmt.Errorf("building phase state: %w", err)
	}

	winnerUserID := ""
	if snap.Game.WinnerUserID.Valid {
		winnerUserID = snap.Game.WinnerUserID.String
	}

	return struct {
		ID     int64  `json:"id"`
		Turn   int64  `json:"turn"`
		Phase  any    `json:"phase"`
		Winner string `json:"winnerUserId"`
	}{
		ID:   snap.Game.ID,
		Turn: snap.Game.Turn,
		Phase: struct {
			Type  game.PhaseType `json:"type"`
			State any            `json:"state"`
		}{
			Type:  phaseType,
			State: phaseState,
		},
		Winner: winnerUserID,
	}, nil
}

// buildPhaseState returns the phase-specific state payload, matching what
// PhaseController produces for each phase type.
func buildPhaseState(phase snapshot.PhaseState) (any, error) {
	switch phase.Type {
	case sqlc.GamePhaseTypeDEPLOY:
		if phase.DeployState == nil {
			return nil, errors.New("deploy phase state is nil")
		}

		return messaging.DeployPhaseState{
			DeployableTroops: phase.DeployState.DeployableTroops,
		}, nil
	case sqlc.GamePhaseTypeCONQUER:
		if phase.ConquerState == nil {
			return nil, errors.New("conquer phase state is nil")
		}

		return messaging.ConquerPhaseState{
			AttackingRegionID: phase.ConquerState.SourceRegion,
			DefendingRegionID: phase.ConquerState.TargetRegion,
			MinTroopsToMove:   phase.ConquerState.MinimumTroops,
		}, nil
	case sqlc.GamePhaseTypeATTACK, sqlc.GamePhaseTypeREINFORCE, sqlc.GamePhaseTypeCARDS:
		return messaging.EmptyState{}, nil
	default:
		return nil, fmt.Errorf("unknown phase type: %s", phase.Type)
	}
}

func buildBoardState(regions []sqlc.GetRegionsByGameRow) messaging.BoardState {
	return messaging.BoardState{Regions: convertRegions(regions)}
}

func convertRegions(regions []sqlc.GetRegionsByGameRow) []messaging.Region {
	result := make([]messaging.Region, len(regions))
	for idx, region := range regions {
		result[idx] = messaging.Region{
			ID:      region.ExternalReference,
			OwnerID: region.UserID,
			Troops:  region.Troops,
		}
	}

	return result
}

func buildPlayerState(
	players []sqlc.GetPlayersStateRow,
	connectedPlayers []string,
) messaging.PlayersState {
	return messaging.PlayersState{
		Players: convertPlayers(players, connectedPlayers),
	}
}

func convertPlayers(
	players []sqlc.GetPlayersStateRow,
	connectedPlayers []string,
) []messaging.Player {
	result := make([]messaging.Player, len(players))
	for idx, player := range players {
		result[idx] = messaging.Player{
			UserID:           player.UserID,
			Name:             player.Name,
			Index:            player.TurnIndex,
			CardCount:        player.CardCount,
			Status:           playerStatus(player.RegionCount),
			ConnectionStatus: connectionStatus(slices.Contains(connectedPlayers, player.UserID)),
		}
	}

	return result
}

func playerStatus(regionCount int64) messaging.PlayerStatus {
	if regionCount == 0 {
		return messaging.Dead
	}

	return messaging.Alive
}

func connectionStatus(isConnected bool) messaging.ConnectionStatus {
	if isConnected {
		return messaging.Connected
	}

	return messaging.Disconnected
}
