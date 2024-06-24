package validation

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
	"go.uber.org/zap"
)

type Service interface {
	Validate(ctx riskcontext.MoveContext, querier db.Querier, game *sqlc.Game) error
}

type ServiceImpl struct {
	log           *zap.SugaredLogger
	playerService player.Service
}

func NewService(
	log *zap.SugaredLogger,
	playerService player.Service,
) *ServiceImpl {
	return &ServiceImpl{log: log, playerService: playerService}
}

func (s *ServiceImpl) Validate(
	ctx riskcontext.MoveContext,
	querier db.Querier,
	game *sqlc.Game,
) error {
	players, err := s.playerService.GetPlayersQ(ctx, querier, game.ID)
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

	return nil
}

func (s *ServiceImpl) checkTurn(
	game *sqlc.Game,
	playersInGame int64,
	playerTurn int64,
) error {
	if game.Turn%playersInGame != playerTurn {
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
