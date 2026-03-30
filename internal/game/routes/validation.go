package routes

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

// ValidateGameWSConnection verifies that the game exists and the requesting
// user is a participant. Returns a ForbiddenError if the user is not in the
// game, or wraps any service errors.
func ValidateGameWSConnection(
	gameCtx ctx.GameContext,
	gameStateService state.Service,
	playerService player.Service,
) error {
	_, err := gameStateService.GetGameState(gameCtx)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	players, err := playerService.GetPlayersState(gameCtx)
	if err != nil {
		return fmt.Errorf("failed to get player state: %w", err)
	}

	if !userIsParticipating(gameCtx.UserID(), players) {
		return domainerrors.NewForbiddenError("user not in game")
	}

	return nil
}

func userIsParticipating(userID string, players []sqlc.GetPlayersStateRow) bool {
	for _, p := range players {
		if p.UserID == userID {
			return true
		}
	}

	return false
}
