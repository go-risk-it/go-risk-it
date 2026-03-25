package validation

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type Service interface {
	Validate(ctx ctx.GameContext, querier db.Querier, game *state.Game) error
}

type service struct {
	playerService player.Service
}

var _ Service = (*service)(nil)

func New(playerService player.Service) Service {
	return &service{playerService: playerService}
}

func (s *service) Validate(ctx ctx.GameContext, querier db.Querier, game *state.Game) error {
	slog.InfoContext(ctx, "performing generic move validation")

	if game.WinnerUserID != "" {
		return domainerrors.NewConflictError("game is already over")
	}

	players, err := s.playerService.GetPlayers(ctx, querier)
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	thisPlayer := extractPlayerFrom(players, ctx.UserID())
	if thisPlayer == nil {
		return domainerrors.NewForbiddenError("player is not in game")
	}

	if err := s.checkTurn(game, int64(len(players)), thisPlayer.TurnIndex); err != nil {
		return fmt.Errorf("turn check failed: %w", err)
	}

	slog.InfoContext(ctx, "generic move validation passed")

	return nil
}

func (s *service) checkTurn(
	game *state.Game,
	playersInGame int64,
	playerTurn int64,
) error {
	if game.Turn%playersInGame != playerTurn {
		return domainerrors.NewConflictError("it is not the player's turn")
	}

	return nil
}

func extractPlayerFrom(players []sqlc.GamePlayer, userID string) *sqlc.GamePlayer {
	for _, p := range players {
		if p.UserID == userID {
			return &p
		}
	}

	return nil
}
