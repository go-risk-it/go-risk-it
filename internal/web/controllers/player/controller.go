package player

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/api/game/message"
	"github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"go.uber.org/zap"
)

type Controller interface {
	GetPlayerState(ctx context.Context, gameID int64) (message.PlayersState, error)
}

type ControllerImpl struct {
	log           *zap.SugaredLogger
	playerService player.Service
}

func New(
	log *zap.SugaredLogger,
	playerService player.Service,
) *ControllerImpl {
	return &ControllerImpl{
		log:           log,
		playerService: playerService,
	}
}

func (c *ControllerImpl) GetPlayerState(
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
		ID:    player.UserID,
		Index: player.TurnIndex,
	}
}
