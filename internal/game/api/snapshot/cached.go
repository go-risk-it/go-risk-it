package snapshot

// CachedGameState holds a fully-mapped game state snapshot suitable for
// immediate serialization to WebSocket clients. The Turn field acts as
// the version stamp for the store's turn-monotonicity guard.
type CachedGameState struct {
	Turn             int64
	ConqueredInTurn  bool
	PublicSnapshot   *GameSnapshot
	PrivateSnapshots map[string]*PlayerPrivate
}
