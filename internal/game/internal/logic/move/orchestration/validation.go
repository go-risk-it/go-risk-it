package orchestration

import (
	"fmt"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/state"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

type ValidationService interface {
	Validate(ctx gamectx.GameContext, querier db.Querier, game *state.Game) error
}

type validationServiceImpl struct {
	playerService player.Service
}

var _ ValidationService = (*validationServiceImpl)(nil)

func NewValidationService(playerService player.Service) ValidationService {
	return &validationServiceImpl{playerService: playerService}
}

func (s *validationServiceImpl) Validate(
	gameCtx gamectx.GameContext,
	querier db.Querier,
	game *state.Game,
) error {
	if game.WinnerUserID != "" {
		return domainerrors.NewConflictError("game is already over")
	}

	players, err := s.playerService.GetPlayers(gameCtx, querier)
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	thisPlayer := extractPlayerFrom(players, gameCtx.UserID())
	if thisPlayer == nil {
		return domainerrors.NewForbiddenError("player is not in game")
	}

	if err := s.checkTurn(game, int64(len(players)), thisPlayer.TurnIndex); err != nil {
		return fmt.Errorf("turn check failed: %w", err)
	}

	return nil
}

func (s *validationServiceImpl) checkTurn(
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
