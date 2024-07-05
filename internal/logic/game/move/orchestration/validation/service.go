package validation

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type Service interface {
	Validate(ctx ctx.MoveContext, querier db.Querier, game *state.Game) error
}

type ServiceImpl struct {
	playerService player.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(playerService player.Service) *ServiceImpl {
	return &ServiceImpl{playerService: playerService}
}

func (s *ServiceImpl) Validate(ctx ctx.MoveContext, querier db.Querier, game *state.Game) error {
	ctx.Log().Infow("performing generic move validation")

	players, err := s.playerService.GetPlayersQ(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	thisPlayer := extractPlayerFrom(players, ctx.UserID())
	if thisPlayer == nil {
		return fmt.Errorf("player is not in game")
	}

	if err := s.checkTurn(game, int64(len(players)), thisPlayer.TurnIndex); err != nil {
		return fmt.Errorf("turn check failed: %w", err)
	}

	ctx.Log().Infow("generic move validation passed")

	return nil
}

func (s *ServiceImpl) checkTurn(
	game *state.Game,
	playersInGame int64,
	playerTurn int64,
) error {
	if game.CurrentTurn%playersInGame != playerTurn {
		return fmt.Errorf("it is not the player's turn")
	}

	return nil
}

func extractPlayerFrom(players []sqlc.Player, userID string) *sqlc.Player {
	for _, p := range players {
		if p.UserID == userID {
			return &p
		}
	}

	return nil
}
