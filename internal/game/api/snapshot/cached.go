package snapshot

// CachedGameState holds a fully-mapped game state snapshot suitable for
// immediate serialization to WebSocket clients. The Turn field acts as
// the version stamp for the store's turn-monotonicity guard. Seq is a
// globally monotonic counter incremented on every Store() call — it
// uniquely identifies the snapshot version across all moves within a game,
// even when multiple moves share the same Turn.
type CachedGameState struct {
	Turn             int64
	Seq              int64
	ConqueredInTurn  bool
	PublicSnapshot   *GameSnapshot
	PrivateSnapshots map[string]*PlayerPrivate
}
