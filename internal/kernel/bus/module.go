package bus

import "go.uber.org/fx"

// Module provides the event bus and its directional sub-interfaces as fx dependencies.
// Consumer packages should depend on Publisher or Subscriber, not Bus.
var Module = fx.Options(
	fx.Provide(NewBus),
	fx.Provide(func(b Bus) Publisher { return b }),
	fx.Provide(func(b Bus) Subscriber { return b }),
)
