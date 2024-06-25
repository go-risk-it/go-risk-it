package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/message"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
)

type PlayerController interface {
	GetPlayerState(ctx ctx.GameContext) (message.PlayersState, error)
}

type PlayerControllerImpl struct {
	playerService player.Service
}

func NewPlayerController(playerService player.Service) *PlayerControllerImpl {
	return &PlayerControllerImpl{playerService: playerService}
}

func (c *PlayerControllerImpl) GetPlayerState(ctx ctx.GameContext) (message.PlayersState, error) {
	ctx.Log().Infow("fetching players")

	players, err := c.playerService.GetPlayers(ctx)
	if err != nil {
		return message.PlayersState{}, fmt.Errorf("unable to get players: %w", err)
	}

	ctx.Log().Infow("got players", "players", players)

	return message.PlayersState{Players: convertPlayers(players)}, nil
}

func convertPlayers(players []sqlc.Player) []message.Player {
	result := make([]message.Player, len(players))
	for i, p := range players {
		result[i] = convertPlayer(p)
	}

	return result
}

func convertPlayer(player sqlc.Player) message.Player {
	return message.Player{
		UserID: player.UserID,
		Name:   player.Name,
		Index:  player.TurnIndex,
	}
}
