package ctx

import "log/slog"

// LogEnricher is implemented by context types that can produce structured slog
// attributes from their domain metadata. The slog ContextHandler uses a single
// LogEnricher type assertion instead of switching on each concrete context type.
// Implementations compose: GameContext.SlogAttrs() delegates to the embedded
// UserContext first, then appends its own gameID attribute.
type LogEnricher interface {
	SlogAttrs() []slog.Attr
}
