package converter

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	game "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

// ConvertPublicSnapshot transforms a PublicSnapshot into pre-serialized WS
// messages for the broadcast path: gameState, boardState, playerState.
//
// connectedPlayers is the list of user IDs currently connected via WebSocket,
// used to enrich each player's ConnectionStatus field.
func ConvertPublicSnapshot(
	snap *snapshot.PublicSnapshot,
	connectedPlayers []string,
) (*PublicMessages, error) {
	gameMsg, err := buildGameStateMessage(snap)
	if err != nil {
		return nil, fmt.Errorf("building game state message: %w", err)
	}

	boardMsg, err := buildBoardStateMessage(snap.Board)
	if err != nil {
		return nil, fmt.Errorf("building board state message: %w", err)
	}

	playerMsg, err := buildPlayerStateMessage(snap.Players, connectedPlayers)
	if err != nil {
		return nil, fmt.Errorf("building player state message: %w", err)
	}

	return &PublicMessages{
		GameState:   gameMsg,
		BoardState:  boardMsg,
		PlayerState: playerMsg,
	}, nil
}

// buildGameStateMessage produces the GameState message envelope.
// The phase-discriminated generic GameState[T] from messaging cannot be
// used here because Go generics require compile-time type parameters.
// Instead we build the same JSON structure using a local untyped struct,
// which produces byte-identical output.
func buildGameStateMessage(snap *snapshot.PublicSnapshot) (json.RawMessage, error) {
	phaseType, err := convertPhaseType(snap.Phase.Type)
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

	gameState := struct {
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
	}

	return message.BuildMessage(message.GameState, gameState)
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

func buildBoardStateMessage(
	regions []sqlc.GetRegionsByGameRow,
) (json.RawMessage, error) {
	boardState := messaging.BoardState{Regions: convertRegions(regions)}

	return message.BuildMessage(message.BoardState, boardState)
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

func buildPlayerStateMessage(
	players []sqlc.GetPlayersStateRow,
	connectedPlayers []string,
) (json.RawMessage, error) {
	playersState := messaging.PlayersState{
		Players: convertPlayers(players, connectedPlayers),
	}

	return message.BuildMessage(message.PlayerState, playersState)
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
