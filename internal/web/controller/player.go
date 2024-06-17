package controller

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/message"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/player"
	"go.uber.org/zap"
)

type PlayerController interface {
	GetPlayerState(ctx context.Context, gameID int64) (message.PlayersState, error)
}

type PlayerControllerImpl struct {
	log           *zap.SugaredLogger
	playerService player.Service
}

func NewPlayerController(
	log *zap.SugaredLogger,
	playerService player.Service,
) *PlayerControllerImpl {
	return &PlayerControllerImpl{
		log:           log,
		playerService: playerService,
	}
}

func (c *PlayerControllerImpl) GetPlayerState(
	ctx context.Context, gameID int64,
) (message.PlayersState, error) {
	players, err := c.playerService.GetPlayers(ctx, gameID)
	if err != nil {
		return message.PlayersState{}, fmt.Errorf("unable to get players: %w", err)
	}

	c.log.Infow("got players", "players", players)

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
