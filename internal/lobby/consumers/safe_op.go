package consumers

import (
	"context"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	lobbyctx "github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
)

// LobbySafeOp wraps bus.SafeOp with typed LobbyContext propagation.
// The action receives a LobbyContext directly — no type assertion needed at call sites.
//
// Rebaseable.Rebase contract guarantees LobbyContext after auto-rebase.
//
//nolint:forcetypeassert
func LobbySafeOp(
	lobbyCtx lobbyctx.LobbyContext,
	name string,
	action func(lobbyctx.LobbyContext) error,
) {
	eventbus.SafeOp(lobbyCtx, name, func(ctx context.Context) error {
		return action(ctx.(lobbyctx.LobbyContext))
	})
}
