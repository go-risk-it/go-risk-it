package logger

import "go.uber.org/fx"

// Module registers the game event logger as an fx consumer.
var Module = fx.Options(fx.Invoke(Register))
