package ctx

// Scoped is implemented by context types that carry a numeric scope identifier
// (e.g., a game ID or lobby ID). It enables generic components — such as the
// WebSocket Manager[C Scoped] — to work with any domain context without
// type-switching on concrete types.
type Scoped interface {
	ScopeID() int64
}
