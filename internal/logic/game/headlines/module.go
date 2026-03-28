package headlines

import "go.uber.org/fx"

// Module registers the headline detector as an fx consumer.
var Module = fx.Options(fx.Invoke(RegisterDetector))
