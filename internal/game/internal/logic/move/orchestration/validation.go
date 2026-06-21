package orchestration

import (
	"fmt"

	apisnapshot "github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/state"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

type ValidationService interface {
	Validate(
		ctx gamectx.GameContext,
		game *state.Game,
		players []apisnapshot.PlayerState,
	) error
}

type validationServiceImpl struct{}

var _ ValidationService = (*validationServiceImpl)(nil)

func NewValidationService() ValidationService {
	return &validationServiceImpl{}
}

func (s *validationServiceImpl) Validate(
	gameCtx gamectx.GameContext,
	game *state.Game,
	players []apisnapshot.PlayerState,
) error {
	if game.WinnerUserID != "" {
		return domainerrors.NewConflictError("game is already over")
	}

	thisPlayer := extractPlayerFrom(players, gameCtx.UserID())
	if thisPlayer == nil {
		return domainerrors.NewForbiddenError("player is not in game")
	}

	if err := checkTurn(game, int64(len(players)), thisPlayer.Index); err != nil {
		return fmt.Errorf("turn check failed: %w", err)
	}

	return nil
}

func checkTurn(
	game *state.Game,
	playersInGame int64,
	playerTurn int64,
) error {
	if game.Turn%playersInGame != playerTurn {
		return domainerrors.NewConflictError("it is not the player's turn")
	}

	return nil
}

func extractPlayerFrom(
	players []apisnapshot.PlayerState,
	userID string,
) *apisnapshot.PlayerState {
	for _, p := range players {
		if p.UserID == userID {
			return &p
		}
	}

	return nil
}
