package consumers

import (
	"context"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
)

// GameSafeOp wraps bus.SafeOp with typed GameContext propagation.
// The action receives a GameContext directly — no type assertion needed at call sites.
//
// Rebaseable.Rebase contract guarantees GameContext after auto-rebase.
//
//nolint:forcetypeassert
func GameSafeOp(
	gameCtx gamectx.GameContext,
	name string,
	action func(gamectx.GameContext) error,
) {
	eventbus.SafeOp(gameCtx, name, func(ctx context.Context) error {
		return action(ctx.(gamectx.GameContext))
	})
}
