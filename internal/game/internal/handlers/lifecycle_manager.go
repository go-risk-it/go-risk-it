package handlers

import (
	gameapi "github.com/go-risk-it/go-risk-it/internal/game/api"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/safego"
	"go.uber.org/fx"
)

// lifecycleManager listens for MoveCompleted events where GameOver is true and
// cleans up the associated scoped resources (WS connection group and state cache).
type lifecycleManager struct {
	scopeLifecycle gameapi.ScopeLifecycle
	stateStore     gameapi.StateStore
}

// LifecycleManagerParams holds the dependencies for the lifecycle manager handler.
type LifecycleManagerParams struct {
	fx.In

	Sub            eventbus.Subscriber
	ScopeLifecycle gameapi.ScopeLifecycle
	StateStore     gameapi.StateStore
}

// RegisterLifecycleManager subscribes the lifecycle manager to MoveCompleted events.
func RegisterLifecycleManager(params LifecycleManagerParams) {
	m := &lifecycleManager{
		scopeLifecycle: params.ScopeLifecycle,
		stateStore:     params.StateStore,
	}

	gameevt.OnGameEvent[*gameevt.MoveCompleted](params.Sub, m.handleMoveCompleted)
}

func (m *lifecycleManager) handleMoveCompleted(
	gameCtx gamectx.GameContext,
	event *gameevt.MoveCompleted,
) {
	if !event.GameOver {
		return
	}

	safego.TypedSafeOp(gameCtx, "lifecycle.cleanup", func(_ gamectx.GameContext) error {
		m.scopeLifecycle.RemoveScope(event.GameID())
		m.stateStore.Remove(event.GameID())

		return nil
	})
}
