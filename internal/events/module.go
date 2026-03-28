package events

import "go.uber.org/fx"

// Module provides the event bus as an fx dependency.
var Module = fx.Options(fx.Provide(NewBus))
