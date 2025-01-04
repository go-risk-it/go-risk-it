package controller

import (
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
)

type PlayerController interface {
	GetPlayerState(ctx ctx.GameContext) (messaging.PlayersState, error)
}

type PlayerControllerImpl struct {
	connectionManager connection.Manager
	playerService     player.Service
}

var _ PlayerController = (*PlayerControllerImpl)(nil)

func NewPlayerController(
	connectionManager connection.Manager,
	playerService player.Service,
) *PlayerControllerImpl {
	return &PlayerControllerImpl{
		connectionManager: connectionManager,
		playerService:     playerService,
	}
}

func (c *PlayerControllerImpl) GetPlayerState(
	ctx ctx.GameContext,
) (messaging.PlayersState, error) {
	playersState, err := c.playerService.GetPlayersState(ctx)
	if err != nil {
		return messaging.PlayersState{}, fmt.Errorf("unable to get playersState: %w", err)
	}

	connectedPlayers := c.connectionManager.GetConnectedPlayers(ctx)

	return messaging.PlayersState{Players: convertPlayers(playersState, connectedPlayers)}, nil
}

func convertPlayers(
	players []sqlc.GetPlayersStateRow,
	connectedPlayers []string,
) []messaging.Player {
	result := make([]messaging.Player, len(players))
	for i, p := range players {
		result[i] = convertPlayer(
			p,
			convertConnectionStatus(slices.Contains(connectedPlayers, p.UserID)),
		)
	}

	return result
}

func convertPlayer(
	player sqlc.GetPlayersStateRow,
	connectionStatus messaging.ConnectionStatus,
) messaging.Player {
	return messaging.Player{
		UserID:           player.UserID,
		Name:             player.Name,
		Index:            player.TurnIndex,
		CardCount:        player.CardCount,
		Status:           getPlayerStatus(player),
		ConnectionStatus: connectionStatus,
	}
}

func convertConnectionStatus(isConnected bool) messaging.ConnectionStatus {
	if isConnected {
		return messaging.Connected
	} else {
		return messaging.Disconnected
	}
}

func getPlayerStatus(player sqlc.GetPlayersStateRow) messaging.PlayerStatus {
	if player.RegionCount == 0 {
		return messaging.Dead
	} else {
		return messaging.Alive
	}
}
