package handlers

import "go.uber.org/fx"

// Module registers all game event handlers and the in-memory StateStore.
var Module = fx.Options(
	fx.Provide(NewStateStore),

	// Stateless handlers consuming MoveCompleted / PlayerConnected.
	fx.Invoke(RegisterStateBroadcaster),
	fx.Invoke(RegisterHeadlinesDetector),
	fx.Invoke(RegisterLifecycleManager),
	fx.Invoke(RegisterObservabilityHandler),
	fx.Invoke(RegisterPlayerConnectHandler),

	// Cross-module handler consuming lobby/events/CreateGameRequested.
	fx.Invoke(RegisterGameCreationHandler),
)
