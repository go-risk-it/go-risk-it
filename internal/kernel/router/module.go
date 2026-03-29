package router

import "go.uber.org/fx"

// Module provides the kernel router for cross-module command dispatch.
var Module = fx.Options(
	fx.Provide(NewRouter),
)
