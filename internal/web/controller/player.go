package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
)

type PlayerController interface {
	GetPlayerState(ctx ctx.GameContext) (messaging.PlayersState, error)
}

type PlayerControllerImpl struct {
	playerService player.Service
}

var _ PlayerController = (*PlayerControllerImpl)(nil)

func NewPlayerController(playerService player.Service) *PlayerControllerImpl {
	return &PlayerControllerImpl{playerService: playerService}
}

func (c *PlayerControllerImpl) GetPlayerState(
	ctx ctx.GameContext,
) (messaging.PlayersState, error) {
	playersState, err := c.playerService.GetPlayersState(ctx)
	if err != nil {
		return messaging.PlayersState{}, fmt.Errorf("unable to get playersState: %w", err)
	}

	return messaging.PlayersState{Players: convertPlayers(playersState)}, nil
}

func convertPlayers(players []sqlc.GetPlayersStateRow) []messaging.Player {
	result := make([]messaging.Player, len(players))
	for i, p := range players {
		result[i] = convertPlayer(p)
	}

	return result
}

func convertPlayer(player sqlc.GetPlayersStateRow) messaging.Player {
	return messaging.Player{
		UserID:    player.UserID,
		Name:      player.Name,
		Index:     player.TurnIndex,
		CardCount: player.CardCount,
	}
}
