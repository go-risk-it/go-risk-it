package snapshot

import "go.uber.org/fx"

// Module provides the snapshot Service for dependency injection.
var Module = fx.Options(
	fx.Provide(NewService),
)
