package validation

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/game/tracing"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
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
	ctx, span := tracing.StartGameSpan(ctx, "game.move.validate")
	defer span.End()

	slog.DebugContext(ctx, "performing generic move validation")

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

	slog.DebugContext(ctx, "generic move validation passed")

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
